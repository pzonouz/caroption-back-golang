package utils

import "github.com/golang-jwt/jwt/v5"

type AuthClaims struct {
	jwt.RegisteredClaims

	ID      string `json:"userId"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
}

type User struct {
	ID      string
	Email   string
	IsAdmin bool
}
