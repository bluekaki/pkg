package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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

	if true {
		fmt.Println("---------------------- normal ----------------------------")

		form := make(url.Values)
		form.Set("message", "hello world")

		signature, date, err := signer.Generate("TESDUM", auth.MethodGet, "/dummy/stream/echo?"+form.Encode(), nil)
		if err != nil {
			panic(err)
		}

		body, header, _, err := httpclient.Get("http://127.0.0.1:8080/dummy/stream/echo", form,
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

	if false {
		fmt.Println("---------------------- upload file ----------------------------")

		payload := make([][]byte, 3)
		for i := range payload {
			buf := make([]byte, 10<<20)
			io.ReadFull(rand.Reader, buf)
			payload[i] = buf
		}

		hash := sha256.New()
		for i := range payload {
			hash.Write(payload[i])
		}

		digest := hash.Sum(nil)
		fmt.Println(hex.EncodeToString(digest[:]))

		signature, date, err := signer.Generate("TESDUM", auth.MethodPost, "/dummy/upload/A Test File", bytes.Join(payload, nil))
		if err != nil {
			panic(err)
		}

		body, header, _, err := httpclient.PostMultipartFile("http://127.0.0.1:8080/dummy/upload/A Test File", payload,
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
