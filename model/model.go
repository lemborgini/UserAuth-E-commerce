package models

import "gorm.io/gorm"

type UserModel struct {
	gorm.Model

	Username         string
	Email            string
	Password         string
	ProductsUploaded int
}

type category int

const (
	wearing category = iota
	play
)

type ProductModel struct {
	gorm.Model

	Name        string
	Description string
	Category    category
	Price       string
	Seller      string
}
