package jwtauth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	//"github.com/dgrijalva/jwt-go"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var secretKey = []byte{125}

type TokenJwt struct {
	Token string `json:"token"`
}

func GenerateJwt() (string, error) {
	const op = "GenerateJwt issue"
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "todo-app",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	fmt.Printf("Token claims added: %+v\n", claims)
	tokenString, err := claims.SignedString(secretKey)
	if err != nil {
		return "", errors.New(op + " " + err.Error())
	}
	return tokenString, nil
}

func VerifyToken(tokenString string) (bool, error) {
	const op = "verifyToken"
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return false, errors.New(op + " " + err.Error())
	}
	if !token.Valid {
		return false, errors.New(op + " invalid token")
	}
	return true, nil
}

func JwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		jwtToken, err := VerifyToken(tokenString)
		if err != nil || !jwtToken {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
