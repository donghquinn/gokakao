// Package kakaotest provides a fake Kakao API server for tests.
//
// One Server stands in for every Kakao host; point each host of the client
// under test at Server.URL.
package kakaotest

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
)

// Request is a request received by the Server.
type Request struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   []byte
}

// Form parses Body as an application/x-www-form-urlencoded body.
func (r Request) Form() url.Values {
	v, _ := url.ParseQuery(string(r.Body))
	return v
}

// Server is a fake Kakao API server. Requests to routes without a handler
// fail the test.
type Server struct {
	URL string

	tb   testing.TB
	srv  *httptest.Server
	mux  *http.ServeMux
	mu   sync.Mutex
	reqs []Request
}

// NewServer starts a Server that is closed when the test ends.
func NewServer(tb testing.TB) *Server {
	tb.Helper()
	s := &Server{tb: tb, mux: http.NewServeMux()}
	s.srv = httptest.NewServer(http.HandlerFunc(s.serve))
	s.URL = s.srv.URL
	tb.Cleanup(s.srv.Close)
	return s
}

// Handle registers h for pattern, which uses [http.ServeMux] syntax,
// e.g. "GET /v2/user/me".
func (s *Server) Handle(pattern string, h http.HandlerFunc) {
	s.mux.HandleFunc(pattern, h)
}

// HandleJSON registers a handler for pattern that replies with status and
// the JSON body.
func (s *Server) HandleJSON(pattern string, status int, body string) {
	s.Handle(pattern, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	})
}

// HandleFile registers a handler for pattern that replies 200 with the
// JSON file at path, typically under testdata/.
func (s *Server) HandleFile(pattern, path string) {
	s.tb.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		s.tb.Fatalf("kakaotest: %v", err)
	}
	s.HandleJSON(pattern, http.StatusOK, string(b))
}

// Requests returns the requests received so far, in order.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.reqs...)
}

// LastRequest returns the most recent request. It fails the test if none
// was received.
func (s *Server) LastRequest() Request {
	s.tb.Helper()
	reqs := s.Requests()
	if len(reqs) == 0 {
		s.tb.Fatal("kakaotest: no request received")
	}
	return reqs[len(reqs)-1]
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	s.mu.Lock()
	s.reqs = append(s.reqs, Request{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.Query(),
		Header: r.Header.Clone(),
		Body:   body,
	})
	s.mu.Unlock()

	if _, pattern := s.mux.Handler(r); pattern == "" {
		s.tb.Errorf("kakaotest: unexpected request %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"code":-2,"msg":"kakaotest: no handler"}`)
		return
	}
	s.mux.ServeHTTP(w, r)
}
