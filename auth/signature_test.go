package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignature(t *testing.T) {
	assert := assert.New(t)

	instance, err := NewSignature(WithSHA256(), WithSecrets(map[Identifier]Secret{
		"AAASys": "1234567890",
		"BBBSys": "0987654321",
	}))
	assert.Nil(err)

	method := MethodPost
	uri := "/echo?key1=value1&key2=value2"
	body := []byte(`{"payload":"Hello World"}`)

	authorization, date, err := instance.Generate("AAASys", method, uri, body)
	assert.Nil(err)

	t.Log("authorization:", authorization)
	t.Log("date:", date)

	identifier, ok, err := instance.Verify(authorization, date, method, uri, body)
	assert.Nil(err)
	assert.True(ok)

	t.Log("identifier:", identifier)
}
