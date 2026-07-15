package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	// database "user-aurth-project"

	"net/http"
	"user-aurth-project/database"
	models "user-aurth-project/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string
	Password string
}

type claims struct {
	jwt.RegisteredClaims

	Username string
	Email    string
}

func CreateUser(c *gin.Context) {
	var user models.UserModel

	if err := c.BindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	result := database.DB.Create(&user)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(201, user)
}

func DeleteUser(c *gin.Context) {
	type ID struct {
		id int
	}
	var id ID
	c.BindJSON(&id)
	var user models.UserModel
	database.DB.First(&user, id.id)
	database.DB.Delete(&user)
	c.IndentedJSON(200, user)

}

func Login(c *gin.Context) {
	var req loginRequest
	var user models.UserModel
	c.BindJSON(&req)
	database.DB.Where("email = ?", req.Email).First(&user)
	err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "Wrong Password"})
		return
	}

	claim := claims{
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claim,
	)
	tokenString, err := token.SignedString(jwtkey)

	if err != nil {
		c.IndentedJSON(http.StatusPreconditionFailed, gin.H{"message": "Something Went wrong"})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
	})
}

var jwtkey = []byte("supra_secrette_kay")

func GetProfile(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	tokenString := strings.TrimPrefix(authToken, " Bearer ")

	claim := &claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claim,
		func(token *jwt.Token) (any, error) {
			return jwtkey, nil
		},
	)

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid token",
		})
		return
	}
	fmt.Println("Authorization:", authToken)
	fmt.Println("Token String:", tokenString)
	c.JSON(http.StatusOK, claim)
}

func createProduct(c *gin.Context) {
	var user models.UserModel
	var product models.ProductModel

	authtoken := c.GetHeader("Authorization")
	tokenstring := strings.TrimPrefix(authtoken, "Bearer ")
	claim := &claims{}
	_, err := jwt.ParseWithClaims(
		tokenstring,
		claim,
		func(t *jwt.Token) (any, error) {
			return jwtkey, nil
		},
	)
	if err != nil {
		c.IndentedJSON(http.StatusExpectationFailed, gin.H{"message": "I know you are an attacker"})
		return
	}
	c.BindJSON(&product)
	product.Seller = claim.Username
	database.DB.Where("Username = ?", claim.Username).First(&user)
	database.DB.Create(&product)
	user.ProductsUploaded++
	c.IndentedJSON(201, gin.H{"message": "PUBLISHED"})
}

func main() {
	database.Connect()
	database.DB.AutoMigrate(&models.UserModel{}, &models.ProductModel{})
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.POST("/createUser", CreateUser)
	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(200, "index.html", nil)
	})
	router.POST("/createProduct", createProduct)
	router.GET("/profile", GetProfile)
	router.POST("/login", Login)
	router.DELETE("/deleteUser", DeleteUser)
	router.Run(":8000")
}
