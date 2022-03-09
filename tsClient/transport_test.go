package tsClient

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestPost(t *testing.T) {

	body := "article body"
	tests := []struct {
		name            string
		want            string
		wantErr         bool
		mockHandlerFunc http.HandlerFunc
		timeout         time.Duration
	}{
		{
			name: "pass",
			mockHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				cType := r.Header.Get("Content-Type")
				if !strings.Contains(cType, "text/plain") {
					t.Fatalf("want Content-Type contains text/plain")
				}
				defer func(rc io.ReadCloser) {
					err := r.Body.Close()
					if err != nil {
						t.Errorf("could not close reader")
					}
				}(r.Body)

				bytes, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("could not read body")
				}
				if string(bytes) != body {
					t.Fatalf("want body=%s, got %s", body, string(bytes))
				}

				_, err = fmt.Fprintf(w, "OK")
				if err != nil {
					t.Fatalf("could not write to responseWriter")
				}
			},
			want:    "OK",
			wantErr: false,
			timeout: time.Second,
		},
		{
			name: "timeout error",
			mockHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Millisecond * 3)
			},
			wantErr: true,
			timeout: time.Millisecond,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				srv := httptest.NewServer(test.mockHandlerFunc)
				defer srv.Close()
				parse, err := url.Parse(srv.URL)
				if err != nil {
					t.Fatalf("could not parse url")
				}
				got, err := post(parse, strings.NewReader(body), test.timeout)

				if (err != nil) != test.wantErr {
					t.Errorf("post() error = %v, wantErr %v", err, test.wantErr)
					return
				}
				if string(got) != test.want {
					t.Errorf("post() got = %v, want %v", string(got), test.want)
				}
			},
		)
	}
}
