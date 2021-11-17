package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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

	if false {
		fmt.Println("---------------------- normal ----------------------------")

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

	if false {
		fmt.Println("---------------------- panic ----------------------------")

		form := make(url.Values)
		form.Set("message", "panic")

		signature, date, err := signer.Generate("TESDUM", auth.MethodGet, "/dummy/echo?"+form.Encode(), nil)
		if err != nil {
			panic(err)
		}

		body, header, statusCode, err := httpclient.Get("http://127.0.0.1:8080/dummy/echo", form,
			httpclient.WithHeader("Authorization", "cBmhBrwHZ0dM5DJy9TK1"),
			httpclient.WithHeader("Date", date),
			httpclient.WithHeader("Authorization-Proxy", signature),
		)
		if err != nil {
			fmt.Println(statusCode, header.Get("journal-id"), err)

		} else {
			fmt.Println(header.Get("journal-id"), string(body))
		}
	}

	if false {
		fmt.Println("---------------------- business err ----------------------------")

		form := make(url.Values)
		form.Set("message", "business err")

		signature, date, err := signer.Generate("TESDUM", auth.MethodGet, "/dummy/echo?"+form.Encode(), nil)
		if err != nil {
			panic(err)
		}

		body, header, statusCode, err := httpclient.Get("http://127.0.0.1:8080/dummy/echo", form,
			httpclient.WithHeader("Authorization", "cBmhBrwHZ0dM5DJy9TK1"),
			httpclient.WithHeader("Date", date),
			httpclient.WithHeader("Authorization-Proxy", signature),
		)
		if err != nil {
			fmt.Println(statusCode, header.Get("journal-id"), err)

		} else {
			fmt.Println(header.Get("journal-id"), string(body))
		}
	}

	if false {
		fmt.Println("---------------------- alert err ----------------------------")

		form := make(url.Values)
		form.Set("message", "alert err")

		signature, date, err := signer.Generate("TESDUM", auth.MethodGet, "/dummy/echo?"+form.Encode(), nil)
		if err != nil {
			panic(err)
		}

		body, header, statusCode, err := httpclient.Get("http://127.0.0.1:8080/dummy/echo", form,
			httpclient.WithHeader("Authorization", "cBmhBrwHZ0dM5DJy9TK1"),
			httpclient.WithHeader("Date", date),
			httpclient.WithHeader("Authorization-Proxy", signature),
		)
		if err != nil {
			fmt.Println(statusCode, header.Get("journal-id"), err)

		} else {
			fmt.Println(header.Get("journal-id"), string(body))
		}
	}

	if true {
		fmt.Println("---------------------- upload file ----------------------------")

		payload, err := ioutil.ReadFile("httpclient.go")
		if err != nil {
			panic(err)
		}

		payload = payload[:200]

		digest := sha256.Sum256(payload)
		fmt.Println(hex.EncodeToString(digest[:]))

		signature, date, err := signer.Generate("TESDUM", auth.MethodPost, "/dummy/upload/测试文件", payload)
		if err != nil {
			panic(err)
		}

		body, header, _, err := httpclient.PostMultipartFile("http://127.0.0.1:8080/dummy/upload/测试文件", payload,
			httpclient.WithHeader("Authorization", "cBmhBrwHZ0dM5DJy9TK1"),
			httpclient.WithHeader("Date", date),
			httpclient.WithHeader("Authorization-Proxy", signature),
		)
		if err != nil {
			panic(err)
		}

		fmt.Println(header.Get("journal-id"), string(body))
	}
}
