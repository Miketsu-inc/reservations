package middlewares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
)

type contextKey struct {
	name string
}

var UserIDCtxKey = &contextKey{"UserID"}

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		claims, err := verifyRequest(r, getTokenFromHeader, getTokenFromCookie)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf(err.Error()))
			return
		}

		userID := getUserIdFromClaims(claims)
		if userID == uuid.Nil {
			utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("could not parse jwt claims"))
			return
		}

		ctx = context.WithValue(ctx, UserIDCtxKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CreateJWT(secret []byte, userID uuid.UUID) (string, error) {
	exp_time, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_TIME"))
	if err != nil {
		log.Fatal(err)
	}

	expiration := time.Second * time.Duration(exp_time)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(expiration).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err

	}

	return tokenString, nil
}

func getUserIdFromClaims(claims jwt.MapClaims) uuid.UUID {
	uuidStr, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil
	}

	userID, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil
	}

	return userID
}

func verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func verifyRequest(r *http.Request, findTokenFns ...func(r *http.Request) string) (jwt.MapClaims, error) {
	var tokenString string

	for _, fn := range findTokenFns {
		tokenString = fn(r)
		if tokenString != "" {
			break
		}
	}
	if tokenString == "" {
		return nil, fmt.Errorf("JWT token could not be found")
	}

	return verifyToken(tokenString)
}

func getTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return ""
	}

	return cookie.Value
}

// Rerive token from "Authorization" request header: "Authorization: BEARER T".
func getTokenFromHeader(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}

	return ""
}
