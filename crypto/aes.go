package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"

	"github.com/byepichi/pkg/errors"
)

func chaos(key string) ([]byte, []byte) {
	raw := sha256.Sum256([]byte(key))
	for k := 1; k < 1000000; k++ {
		raw = sha256.Sum256(raw[:])
	}

	nonce := sha256.Sum256(raw[:])
	return raw[:], nonce[:12]
}

// AesGCM256EncryptWithChaos encrypt by aes-gcm-256 with chaos key(s)
func AesGCM256EncryptWithChaos(keys []string, raw []byte) ([]byte, error) {
	for _, key := range keys {
		encKey, nonce := chaos(key)

		block, err := aes.NewCipher(encKey)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		aead, err := cipher.NewGCM(block)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		raw = aead.Seal(nil, nonce, raw, nil)
	}

	return raw, nil
}

// AesGCM256DecryptWithChaos decrypt by aes-gcm-256 with chaos key(s)
func AesGCM256DecryptWithChaos(keys []string, raw []byte) ([]byte, error) {
	for k := len(keys) - 1; k > -1; k-- {
		encKey, nonce := chaos(keys[k])

		block, err := aes.NewCipher(encKey)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		aead, err := cipher.NewGCM(block)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		raw, err = aead.Open(nil, nonce, raw, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return raw, nil
}

// AesGCM128Encrypt encrypt by aes-gcm-128
func AesGCM128Encrypt(key string, raw []byte) ([]byte, error) {
	digest := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(digest[:16])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return aead.Seal(nil, digest[20:], raw, nil), nil
}

// AesGCM128Decrypt decrypt by aes-gcm-128
func AesGCM128Decrypt(key string, raw []byte) ([]byte, error) {
	digest := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(digest[:16])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	raw, err = aead.Open(nil, digest[20:], raw, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return raw, nil
}
