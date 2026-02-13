package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Test_authMiddleware(t *testing.T) {
	a := App{cfg: Config{Server: ServerConfig{JWTHeader: "Auth", JWTKey: "test"}}}

	res := &httptest.ResponseRecorder{}
	req := &http.Request{Header: make(http.Header)}

	// no header
	a.authMiddleware(nil).ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatal(res.Code)
	}

	// wrong header - random jwt token
	resInvalid := &httptest.ResponseRecorder{}
	req.Header.Set(a.cfg.Server.JWTHeader, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30")
	a.authMiddleware(nil).ServeHTTP(resInvalid, req)

	if resInvalid.Code != http.StatusForbidden {
		t.Fatal(resInvalid.Code)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "vfs",
		Subject:   "test",
	})

	key := []byte(a.cfg.Server.JWTKey)
	tokenString, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}

	// expired header
	resExpired := &httptest.ResponseRecorder{}
	req.Header.Set(a.cfg.Server.JWTHeader, tokenString)
	a.authMiddleware(nil).ServeHTTP(resExpired, req)

	if resExpired.Code != http.StatusUnauthorized {
		t.Fatal(resExpired.Code)
	}
}
