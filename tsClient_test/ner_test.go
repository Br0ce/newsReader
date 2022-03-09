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

func TestNERProcess(t *testing.T) {
	body := "some txt"
	pers := []string{"per"}
	locs := []string{"loc"}
	orgs := []string{"org"}
	s := strings.Builder{}
	for i := 0; i < 600; i++ {
		s.WriteString("a")
	}
	invalidBodyLen := s.String()

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
			want:    newsReader.Article{Body: body, Pers: pers, Locs: locs, Orgs: orgs},
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

				type res struct {
					Token string `json:"token"`
					Pred  string `json:"pred"`
				}
				rr := []res{
					{
						Token: "per",
						Pred:  "B-PER",
					},
					{
						Token: "loc",
						Pred:  "B-LOC",
					},
					{
						Token: "org",
						Pred:  "B-ORG",
					},
				}
				bytes, err := json.Marshal(rr)
				if err != nil {
					t.Fatalf("could not marshall article")
				}
				w.WriteHeader(200)
				_, err = fmt.Fprintf(w, string(bytes))
				if err != nil {
					t.Fatalf("could not write to responseWriter")
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
			name:    "invalid response",
			arg:     newsReader.Article{},
			timeout: time.Second,
			wantErr: true,
			tsServer: func(w http.ResponseWriter, r *http.Request) {
				_, err := fmt.Fprintf(w, "")
				if err != nil {
					t.Fatalf("could not write to responseWriter")
				}
			},
		},
		{
			name:    "body length truncated",
			arg:     newsReader.Article{Body: invalidBodyLen},
			timeout: time.Second,
			want: newsReader.Article{
				Body: invalidBodyLen,
				Pers: []string{},
				Locs: []string{},
				Orgs: []string{},
			},
			wantErr: false,
			tsServer: func(w http.ResponseWriter, r *http.Request) {
				defer func(rc io.ReadCloser) {
					err := rc.Close()
					if err != nil {
						t.Fatalf("could not close body")
					}
				}(r.Body)
				bytes, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("could not read body")
				}
				if len(bytes) > 512 {
					t.Fatalf("body len > 512")
				}

				var rr []struct {
					Token string `json:"token"`
					Pred  string `json:"pred"`
				}
				bytes, err = json.Marshal(rr)
				if err != nil {
					t.Fatalf("could not marshal response")
				}
				_, err = fmt.Fprintf(w, string(bytes))
				if err != nil {
					t.Fatalf("could not write to responseWriter")
				}
			},
		},
		{
			name:    "timeout error",
			arg:     newsReader.Article{Body: body},
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
					t.Fatalf("could not parse url")
				}

				ner, err := tsClient.NewNER(u.Host, logger, test.timeout)
				if err != nil {
					t.Fatalf("could not create new NER")
					return
				}

				got, err := ner.Process(test.arg)

				if (err != nil) != test.wantErr {
					t.Fatalf("got error=%v, want=%v", err, test.wantErr)
				}

				if !reflect.DeepEqual(got, test.want) {
					t.Fatalf("want=%v, got=%v", test.want, got)
				}
			},
		)
	}
}
