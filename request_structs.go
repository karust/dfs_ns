package main

import jwt "github.com/dgrijalva/jwt-go"

// Auth ... PSK for storage server authorization
type Auth struct {
	ID uint `json:"id"`
	jwt.StandardClaims
}
