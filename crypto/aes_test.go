package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AESGCM256WithChaos(t *testing.T) {
	assert := assert.New(t)

	keys := []string{"A", "B", "C", "D"}
	message := []byte("hello world !")

	raw, err := AesGCM256EncryptWithChaos(keys, message)
	assert.Nil(err)
	msg, err := AesGCM256DecryptWithChaos(keys, raw)
	assert.Nil(err)

	assert.Equal(message, msg)
}

func Test_AESGCM128(t *testing.T) {
	assert := assert.New(t)

	key := "xxxx"
	message := []byte("hello world !")

	raw, err := AesGCM128Encrypt(key, message)
	assert.Nil(err)
	msg, err := AesGCM128Decrypt(key, raw)
	assert.Nil(err)

	assert.Equal(message, msg)
}
