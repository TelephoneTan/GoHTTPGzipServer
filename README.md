# GoHTTPGzipServer

GoLang `http.Handler` with transparent Gzip support.

## Usage

```go
package main

import (
	"github.com/TelephoneTan/GoHTTPGzipServer/gzip"
	"net/http"
)

func main() {
	h := http.NewServeMux()
	h.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hello, world"))
	})
	gzipH := (&gzip.Handler{Handler: h}).Init()
	http.ListenAndServe("localhost:3000", gzipH)
}

```