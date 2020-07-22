package business

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

type Argon2IdOpts struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultArgon2IdOpts = Argon2IdOpts{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

func hashSecret(salt []byte, secret string, p Argon2IdOpts) (hash string) {
	h := argon2.IDKey([]byte(secret), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(h)

	hash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.Memory, p.Iterations, p.Parallelism, b64Salt, b64Hash)

	return hash
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func compareSecretAndHash(secret, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded secret
	// hash.
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return
	}

	// Derive the key from the other secret using the same parameters.
	secretHash := argon2.IDKey([]byte(secret), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	// Check that the contents of the hashed secrets are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, secretHash) == 1 {
		match = true
	}

	return
}

func decodeHash(encodedHash string) (opts Argon2IdOpts, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		err = ErrInvalidHash

		return
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return
	}
	if version != argon2.Version {
		err = ErrIncompatibleVersion

		return
	}

	opts = Argon2IdOpts{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &opts.Memory, &opts.Iterations, &opts.Parallelism)
	if err != nil {
		return
	}

	salt, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return
	}
	opts.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return
	}
	opts.KeyLength = uint32(len(hash))

	return
}
