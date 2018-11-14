package main

import jwt "github.com/dgrijalva/jwt-go"

// Auth ... PSK for storage server authorization
type Auth struct {
	ID uint `json:"id"`
	jwt.StandardClaims
}

// FileDir ...
type FileDir struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isdir"`
}
