package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const BcryptCost = 12

// --- JWT Claims ---

// AdminClaims for system admin users
type AdminClaims struct {
	AdminID string `json:"admin_id"`
	jwt.RegisteredClaims
}

// UserClaims for backoffice users
type UserClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// AppUserClaims for app users
type AppUserClaims struct {
	AppUserID string `json:"app_user_id"`
	TenantID  string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// --- JWT Token Generation ---

// GenerateAdminToken generates a JWT token for admin
func GenerateAdminToken(adminID, secret string, expiryHours int) (string, error) {
	claims := AdminClaims{
		AdminID: adminID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "admin:" + adminID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateUserToken generates a JWT token for backoffice user
func GenerateUserToken(userID, tenantID, secret string, expiryHours int) (string, error) {
	claims := UserClaims{
		UserID:   userID,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "user:" + userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateAppUserToken generates a JWT token for app user
func GenerateAppUserToken(appUserID, tenantID, secret string, expiryHours int) (string, error) {
	claims := AppUserClaims{
		AppUserID: appUserID,
		TenantID:  tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "app:" + appUserID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// --- JWT Token Validation ---

// ValidateAdminToken validates and parses an admin JWT token
func ValidateAdminToken(tokenString, secret string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ValidateUserToken validates and parses a user JWT token
func ValidateUserToken(tokenString, secret string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ValidateAppUserToken validates and parses an app user JWT token
func ValidateAppUserToken(tokenString, secret string) (*AppUserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AppUserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AppUserClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// --- Password Hashing ---

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	return string(bytes), err
}

// CheckPassword checks a password against a bcrypt hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
