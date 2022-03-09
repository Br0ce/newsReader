package tsClient_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"newsReader"
	"newsReader/tsClient"
)

func TestSummaryProcess(t *testing.T) {

	body := "some txt"
	summary := "a summary"

	tests := []struct {
		name     string
		arg      newsReader.Article
		timeout  time.Duration
		want     newsReader.Article
		wantErr  bool
		tsServer http.HandlerFunc
	}{
		{
			name:    "pass",
			arg:     newsReader.Article{Body: body},
			timeout: time.Second,
			want:    newsReader.Article{Body: body, Summary: summary},
			wantErr: false,
			tsServer: func(w http.ResponseWriter, r *http.Request) {
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

				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("could not read body")
				}
				if string(b) != body {
					t.Fatalf("want body=%s, got %s", body, string(b))
				}
				m := make(map[string]string)
				m["summary"] = summary
				response, err := json.Marshal(m)
				if err != nil {
					t.Fatalf("could not marshal response")
				}
				_, err = fmt.Fprintf(w, string(response))
				if err != nil {
					t.Fatalf("could not write response")
				}
			},
		},
		{
			name:    "header does not match",
			arg:     newsReader.Article{},
			timeout: time.Second,
			wantErr: true,
			tsServer: func(w http.ResponseWriter, r *http.Request) {
				cType := r.Header.Get("Content-Type")
				if !strings.Contains(cType, "text/plain") {
					t.Fatalf("want Content-Type contains text/plain")
				}

				w.WriteHeader(400)
				_, err := fmt.Fprintf(w, "bad request")
				if err != nil {
					t.Fatalf("could not write to responseWriter")
				}
			},
		},
		{
			name:    "server error",
			arg:     newsReader.Article{},
			timeout: time.Second,
			wantErr: true,
			tsServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
				_, err := fmt.Fprintf(w, "")
				if err != nil {
					t.Fatalf("could not write to responseWriter")
				}
			},
		},
		{
			name:    "timeout error",
			arg:     newsReader.Article{},
			timeout: time.Millisecond,
			wantErr: true,
			tsServer: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Millisecond * 3)
			},
		},
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	zapper, err := cfg.Build()
	if err != nil {
		t.Fatalf("could not init logger")
	}

	logger := zapper.Sugar()

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				srv := httptest.NewServer(test.tsServer)
				defer srv.Close()
				u, err := url.Parse(srv.URL)
				if err != nil {
					t.Fatalf("could not parse addr")
				}

				summary, err := tsClient.NewSummary(u.Host, logger, test.timeout)
				if err != nil {
					t.Fatalf("could not create new summary")
					return
				}

				got, err := summary.Process(test.arg)

				if (err != nil) != test.wantErr {
					t.Fatalf("want error=%v, got %v", test.wantErr, err)
				}

				if !reflect.DeepEqual(got, test.want) {
					t.Fatalf("want article=%v, got %v", test.want, got)
				}

			},
		)
	}
}
