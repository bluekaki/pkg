package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/byepichi/pkg/auth"
)

func main() {
	signature, err := auth.NewSignature(
		auth.WithSHA256(),
		auth.WithTTL(time.Minute),
		auth.WithSecrets(map[auth.Identifier]auth.Secret{
			"Adummy": "czvZ1khr0XxLNiu8>v)V=~8toA5LJU",
		}))
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		method := req.Method
		uri := req.RequestURI
		proxyAuthorization := req.Header.Get("proxy-authorization")
		date := req.Header.Get("date")
		body, _ := ioutil.ReadAll(req.Body)

		fmt.Println("method:", method)
		fmt.Println("uri:", uri)
		fmt.Println("proxy-authorization:", proxyAuthorization)
		fmt.Println("date:", date)
		fmt.Println("body:", string(body))

		identifier, ok, err := signature.Verify(proxyAuthorization, date, auth.ToMethod(method), uri, body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		fmt.Println("identifier:", identifier)

		io.WriteString(w, "ok")
	})

	if err = http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
