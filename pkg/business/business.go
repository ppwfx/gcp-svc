package business

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/ppwfx/user-svc/pkg/types"
	"net/http"
	"strings"
	"time"
)

func HashPassword(salt string, password string) string {
	return fmt.Sprintf("%x", sha1.New().Sum([]byte(salt+password)))
}

func IsAuthorized(hmacSecret string, accessToken string, allowedSubjectSuffix string) (is bool, err error) {
	t, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(hmacSecret), nil
	})
	if err != nil {
		return
	}

	c, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("expected jwt.StandardClaims")

		return
	}

	sub, ok := c[types.ClaimSub].(string)
	if !ok {
		err = errors.New("expected sub claim")

		return
	}

	if strings.HasSuffix(sub, allowedSubjectSuffix) {
		is = true
	}

	return
}

func ExtractAccessToken(r *http.Request) (t string) {
	return strings.TrimPrefix(r.Header.Get(types.HeaderAuthorization), types.PrefixBearer)
}

func GenerateAccessToken(hmacSecret string, email string) (t string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		Subject:   email,
	})

	t, err = token.SignedString([]byte(hmacSecret))
	if err != nil {
		return
	}

	return
}
