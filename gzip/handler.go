package gzip

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Handler struct {
	headerForbidden     bool
	contentForbidden    bool
	contentEncodingSent bool
	responseWriter      http.ResponseWriter
	gzipWriter          *gzip.Writer
	Handler             http.Handler
}

func (h *Handler) Init() *Handler {
	return h
}

func (h *Handler) removeContentLength() {
	h.Header().Del("Content-Length")
}

func (h *Handler) Header() http.Header {
	return h.responseWriter.Header()
}

func (h *Handler) Write(bs []byte) (num int, err error) {
	if h.contentForbidden {
		return 0, errors.New("content forbidden")
	}
	if len(bs) > 0 {
		h.removeContentLength()
		num, err = h.gzipWriter.Write(bs)
		h.headerForbidden = true
		h.contentEncodingSent = true
	}
	return num, err
}

func (h *Handler) WriteHeader(statusCode int) {
	if h.headerForbidden {
		return
	} else {
		h.headerForbidden = true
	}
	// 如果用户调用此方法并传入这些状态码，那么按照 HTTP 规范，HTTP 回复不应包含回复体。详见：
	//
	// https://www.rfc-editor.org/rfc/rfc9110#name-informational-1xx
	//
	// https://www.rfc-editor.org/rfc/rfc9110#name-204-no-content
	//
	// https://www.rfc-editor.org/rfc/rfc9110#name-205-reset-content
	if statusCode < 200 || statusCode == 204 || statusCode == 205 {
		h.contentForbidden = true
		h.Header().Del("Content-Encoding")
	} else {
		h.contentEncodingSent = true
	}
	h.removeContentLength()
	h.responseWriter.WriteHeader(statusCode)
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

func containsOrSubStringIgnoreCase(list []string, s string) bool {
	s = strings.ToLower(s)
	for _, m := range list {
		m = strings.ToLower(m)
		if m == s || strings.Contains(m, s) {
			return true
		}
	}
	return false
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	aes := r.Header.Values("Accept-Encoding")
	if !containsOrSubStringIgnoreCase(aes, "gzip") {
		h.Handler.ServeHTTP(w, r)
		return
	}
	//
	hCopy := *h
	//goland:noinspection GoAssignmentToReceiver
	h = &hCopy
	//
	h.responseWriter = w
	//
	gw := gzipWriterPool.Get().(*gzip.Writer)
	gw.Reset(w)
	h.gzipWriter = gw
	//
	w.Header().Set("Content-Encoding", "gzip")
	//
	r.Header.Del("Accept-Encoding")
	//
	h.Handler.ServeHTTP(h, r)
	//
	for _, ae := range aes {
		r.Header.Add("Accept-Encoding", ae)
	}
	//
	h.removeContentLength()
	//
	if !h.contentEncodingSent {
		w.Header().Del("Content-Encoding")
	} else {
		_ = h.gzipWriter.Close()
	}
	//
	h.gzipWriter.Reset(io.Discard)
	gzipWriterPool.Put(h.gzipWriter)
}

func (h *Handler) Hijack() (conn net.Conn, readWriter *bufio.ReadWriter, err error) {
	switch w := h.responseWriter.(type) {
	case http.Hijacker:
		return w.Hijack()
	default:
		return conn, readWriter, errors.New("http.ResponseWriter does not implement http.Hijacker")
	}
}
