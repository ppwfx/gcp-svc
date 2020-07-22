package business

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/ppwfx/user-svc/pkg/types"
)

func GetJwtClaims(hmacSecret string, accessToken string) (c map[string]interface{}, err error) {
	t, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(hmacSecret), nil
	})
	if err != nil {
		return
	}
	if !t.Valid {
		err = errors.New("jwt token is not valid")

		return
	}

	c, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("expected jwt.MapClaims")

		return
	}

	return
}

func GenerateAccessToken(hmacSecret string, role string, userID int) (t string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		types.ClaimIat:  time.Now().Unix(),
		types.ClaimExp:  time.Now().Add(24 * time.Hour).Unix(),
		types.ClaimRole: role,
		types.ClaimSub:  userID,
	})

	t, err = token.SignedString([]byte(hmacSecret))
	if err != nil {
		return
	}

	return
}
