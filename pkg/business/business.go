package business

import (
	"crypto/sha1"
	"fmt"
)

func HashPassword(salt string, password string) string {
	return fmt.Sprintf("%x", sha1.New().Sum([]byte(salt+password)))
}
