package main

import (
	"fmt"
	"net/url"

	"github.com/bluekaki/pkg/auth"
	"github.com/bluekaki/pkg/mm/httpclient"
)

func main() {
	signer, err := auth.NewSignature(auth.WithSHA256(), auth.WithSecrets(
		map[auth.Identifier]auth.Secret{
			"TESDUM": "9VbN+~_+8*,9WJ}#}^ZaoW)0=E>AaK",
		}),
	)
	if err != nil {
		panic(err)
	}

	form := make(url.Values)
	form.Set("message", "hello world")

	signature, date, err := signer.Generate("TESDUM", auth.MethodGet, "/dummy/echo?"+form.Encode(), nil)
	if err != nil {
		panic(err)
	}

	body, header, _, err := httpclient.Get("http://127.0.0.1:8080/dummy/echo", form,
		httpclient.WithHeader("Authorization", "cBmhBrwHZ0dM5DJy9TK1"),
		httpclient.WithHeader("Date", date),
		httpclient.WithHeader("Authorization-Proxy", signature),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(header.Get("journal-id"), string(body))
}
