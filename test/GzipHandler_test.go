package test

import (
	"github.com/TelephoneTan/GoHTTPGzipServer/gzip"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func TestGzipHandler(t *testing.T) {
	startServer()
	//
	println("== Gzipped Request =======")
	request(true)
	//
	println("== Not Gzipped Request =======")
	request(false)
}

func request(gzip bool) {
	header := http.Header{}
	if !gzip {
		header.Set("Accept-Encoding", "identity")
	}
	u, _ := url.Parse("http://localhost:3000")
	req := &http.Request{
		Method: http.MethodGet,
		URL:    u,
		Header: header,
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		println("request error:", err.Error())
		return
	}
	text, err := io.ReadAll(response.Body)
	if err != nil {
		println("read error:", err.Error())
		return
	}
	_ = response.Body.Close()
	println("-- Response -------")
	println(string(text))
	println("#############")
	_ = response.Header.Write(os.Stdout)
}

func startServer() {
	helloWorldH := http.NewServeMux()
	helloWorldH.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hello, world"))
	})
	gzipH := (&gzip.Handler{Handler: helloWorldH}).Init()
	printHeaderH := http.NewServeMux()
	printHeaderH.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_ = request.Header.Write(os.Stdout)
		gzipH.ServeHTTP(writer, request)
	})
	go http.ListenAndServe("localhost:3000", printHeaderH)
}
