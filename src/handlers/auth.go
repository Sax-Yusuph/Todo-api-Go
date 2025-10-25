package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sax-yusuph/todo/models"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

type Auth struct {
	Repo *redis.Client
}

// Claims is a custom struct for JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Register handleRegister handles POST /register
func (a *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if u.Username == "" || u.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Check if user already exists
	exists, err := a.Repo.HExists(r.Context(), "users", u.Username).Result()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		respondWithError(w, http.StatusConflict, "Username already taken")
		return
	}

	hashedPassword, _ := hashPassword(u.Password)

	// Store user in 'users' hash
	if err := a.Repo.HSet(r.Context(), "users", u.Username, hashedPassword).Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"message": "User registered"})
}

// Login handleLogin handles POST /login
func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get hashed password from Redis
	hashedPassword, err := a.Repo.HGet(r.Context(), "users", u.Username).Result()
	if errors.Is(err, redis.Nil) {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Check password
	if !checkPasswordHash(u.Password, hashedPassword) {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	tokenString, err := generateJWT(u.Username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

// Middleware AuthMiddleware protects routes
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get Token from Header (e.g., Authorization: Bearer <token>)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			respondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Validate Token & Extract User ID
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Ensure token is signed with HS256 algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			if errors.Is(err, jwt.ErrTokenExpired) {
				respondWithError(w, http.StatusUnauthorized, "Token expired")
				return
			}
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// This is the authenticated user's ID, extracted from the valid token
		tokenUserID := claims.Username

		// 3. Authorize Request (Scope Check)
		// Get the user ID from the URL path parameter
		urlUserID := chi.URLParam(r, "userID")

		if tokenUserID != urlUserID {
			// CRITICAL SECURITY CHECK: Prevents userB from accessing userA's data
			respondWithError(w, http.StatusForbidden, "Access denied: Token user does not match URL user")
			return
		}

		// Optional: Add the authenticated user ID to the request context
		ctx := context.WithValue(r.Context(), "userID", tokenUserID)

		// All good, pass to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to hash a password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Helper to check a password
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	return tokenString, err
}
