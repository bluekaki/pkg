package captcha

import (
	"encoding/base64"
	"testing"
)

func TestCaptcha(t *testing.T) {
	captcha, err := New(200, 80)
	if err != nil {
		t.Fatal(err)
	}

	// show img https://base64.guru/converter/decode/image/png
	code, img := captcha.Simple()
	t.Log(code)
	t.Log(base64.StdEncoding.EncodeToString(img))
}
