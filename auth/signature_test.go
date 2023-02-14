package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignature(t *testing.T) {
	assert := assert.New(t)

	instance, err := NewSignature(WithSHA256(), WithSecrets(map[Identifier]Secret{
		"Adummy": "czvZ1khr0XxLNiu8>v)V=~8toA5LJU",
	}))
	assert.Nil(err)

	method := MethodPost
	uri := "/echo?key1=%CE%95%CE%BB%CE%BB%CE%B7%CE%BD%CE%B9%CE%BA%CF%8C&key2=%CE%B1%CE%BB%CF%86%CE%AC%CE%B2%CE%B7%CF%84%CE%BF"
	body := []byte(`{"payload":"Hello World"}`)

	authorization, date, err := instance.Generate("Adummy", method, uri, body)
	assert.Nil(err)

	t.Log("authorization:", authorization)
	t.Log("date:", date)

	identifier, ok, err := instance.Verify(authorization, date, method, uri, body)
	assert.Nil(err)
	assert.True(ok)

	assert.Equal(identifier, "Adummy")
}
