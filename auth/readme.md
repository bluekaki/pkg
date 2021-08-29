# Raw Example
```go
package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	stdurl "net/url"
	"time"
)

func main() {
	const (
		Prefix = "Adummy"
		Secret = "czvZ1khr0XxLNiu8>v)V=~8toA5LJU"
	)

	// querystring
	form := make(stdurl.Values)
	form.Set("desc", "hello world")
	form.Set("nonce", "0987654321")
	fmt.Println(form.Encode()) // urlencode: desc=hello+world&nonce=0987654321

	uri := "/dummy/hello?" + form.Encode() // urlencode
	fmt.Println(uri)                       // /dummy/hello?desc=hello+world&nonce=0987654321

	body := `{"Address":"LA","Memo":"Unknow"}` // json body

	tz, _ := time.LoadLocation("America/Los_Angeles")
	timestamp := time.Date(2022, time.February, 22, 22, 22, 22, 0, tz)
	fmt.Println(timestamp) // 2022-02-22 22:22:22 -0800 PST

	gmt := timestamp.UTC().Format(http.TimeFormat) // GMT(RFC1123)
	fmt.Println(gmt)                               // Wed, 23 Feb 2022 06:22:22 GMT

	payload := "POST" + "|" + uri + "|" + body + "|" + gmt // POST GET PUT DELETE PATCH must be upper
	fmt.Println(payload)                                   // POST|/dummy/hello?desc=hello+world&nonce=0987654321|{"Address":"LA","Memo":"Unknow"}|Wed, 23 Feb 2022 06:22:22 GMT
	// note： if uri or body is empty
	//           uri empty  payload := "POST" + "|" + "" + "|" + body + "|" + gmt
	//          body empty  payload := "POST" + "|" + uri + "|" + "" + "|" + gmt
	// uri body both empty  payload := "POST" + "|" + "" + "|" + "" + "|" + gmt

	hash := hmac.New(sha256.New, []byte(Secret)) // hamc-sha256
	hash.Write([]byte(payload))
	digest := hash.Sum(nil)

	signature := base64.StdEncoding.EncodeToString(digest) // base64
	fmt.Println(signature)                                 // aZscBeWqPhdlpiuNHiV9iemUQWIfnJjThcEFEuvNtZM=

	req, _ := http.NewRequest("POST", "https://xxx.com"+uri, bytes.NewReader([]byte(body))) // a dummy post request
	req.Header.Set("Proxy-Authorization", Prefix+" "+signature)                             // put "a-space" between prefix and signature
	req.Header.Set("Date", gmt)

	// http.DefaultClient.Do(req)
	// ....
}
```