package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sax-yusuph/todo/models"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Repo      *redis.Client
	JWTSecret string
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

	hashedPassword, _ := a.hashPassword(u.Password)

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
	if !a.checkPasswordHash(u.Password, hashedPassword) {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	tokenString, err := a.generateJWT(u.Username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate token, %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

// Logout handles POST /logout - blacklists the token
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse token to get expiration time
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(a.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Calculate time until token expires
	expirationTime := claims.ExpiresAt.Time
	ttl := time.Until(expirationTime)

	// Only blacklist if token hasn't expired yet
	if ttl > 0 {
		// Store token in Redis blacklist with TTL matching token expiration
		blacklistKey := fmt.Sprintf("blacklist:%s", tokenString)
		if err := a.Repo.Set(r.Context(), blacklistKey, "revoked", ttl).Err(); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to logout")
			return
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
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

		// 2. Check if token is blacklisted
		blacklistKey := fmt.Sprintf("blacklist:%s", tokenString)
		exists, err := a.Repo.Exists(r.Context(), blacklistKey).Result()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}
		if exists > 0 {
			respondWithError(w, http.StatusUnauthorized, "Token has been revoked")
			return
		}

		// 3. Validate Token & Extract User ID
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			// Ensure token is signed with HS256 algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
			}

			return []byte(a.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			if errors.Is(err, jwt.ErrTokenExpired) {
				respondWithError(w, http.StatusUnauthorized, "Token expired")
				return
			}
			respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid token %v", err))
			return
		}

		// This is the authenticated user's ID, extracted from the valid token
		tokenUserID := claims.Username

		// 4. Authorize Request (Scope Check)
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
func (a Auth) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Helper to check a password
func (a Auth) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (a Auth) generateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.JWTSecret))
	return tokenString, err
}
