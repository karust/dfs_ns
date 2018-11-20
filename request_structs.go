package main

import jwt "github.com/dgrijalva/jwt-go"

// Auth ... PSK for storage server authorization
type Auth struct {
	ID uint `json:"id"`
	jwt.StandardClaims
}

// FileDir ...
type FileDir struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	CrTime int64  `json:"time"`
	Size   uint   `json:"size"`
	IsDir  bool   `json:"isdir"`
}
