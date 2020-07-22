// +build unit

package business

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashSecret(t *testing.T) {
	h := hashSecret([]byte("fVWhxUHb0A2CBxHh"), "secret", Argon2IdOpts{
		Memory:      64 * 1024,
		Iterations:  2,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	})

	assert.Equal(t, "$argon2id$v=19$m=65536,t=2,p=2$ZlZXaHhVSGIwQTJDQnhIaA$10bK6t+clO4SfEH2YRL4Q/qg2Vg2cPMEl8snLR5y8ng", h)
}

func TestGenerateRandomBytes(t *testing.T) {
	b, err := generateRandomBytes(16)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, b, 16)
}

func TestCompareSecretAndHash(t *testing.T) {
	salt, err := generateRandomBytes(16)
	if !assert.NoError(t, err) {
		return
	}

	h := hashSecret(salt, "secret", DefaultArgon2IdOpts)

	m, err := compareSecretAndHash("secret", h)
	if !assert.NoError(t, err) {
		return
	}

	assert.True(t, m)

	m, err = compareSecretAndHash("other secret", h)
	if !assert.NoError(t, err) {
		return
	}

	assert.False(t, m)
}
