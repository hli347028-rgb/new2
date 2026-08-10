package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var chinaLocation = mustLoadLocation("Asia/Shanghai")

// ChinaLocation returns Asia/Shanghai timezone.
func ChinaLocation() *time.Location {
	return chinaLocation
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

// NextChinaMidnight returns the next 00:00 in China timezone.
func NextChinaMidnight(now time.Time) time.Time {
	now = now.In(chinaLocation)
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, chinaLocation)
}

type Claims struct {
	Address string `json:"address"`
	jwt.RegisteredClaims
}

// Generate creates a JWT that expires at the next China midnight.
func Generate(address, secret string, now time.Time) (string, time.Time, error) {
	expireAt := NextChinaMidnight(now)
	claims := Claims{
		Address: address,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   address,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expireAt, nil
}

// Parse validates token and returns address.
func Parse(tokenString, secret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", jwt.ErrTokenInvalidClaims
	}
	return claims.Address, nil
}
