package eventStore_test

import (
	"fmt"
	"testing"

	"go.uber.org/zap"
	"newsReader"
	"newsReader/eventStore"
	"newsReader/mock"
)

func TestPublisher(t *testing.T) {
	wantedEType := "type"
	tests := []struct {
		Name string

		PublishFn      func(a newsReader.Article, eType string) error
		PublishInvoked bool

		EType   string
		Article newsReader.Article
		wantErr bool
	}{
		{
			Name: "pass",

			PublishFn: func(a newsReader.Article, eType string) error {
				if a.ID == "" {
					t.Fatalf("Article id must not be empty")
				}
				if eType != wantedEType {
					t.Fatalf("wanted eventType=%v got=%v", wantedEType, eType)
				}
				return nil
			},
			PublishInvoked: true,

			EType:   wantedEType,
			wantErr: false,
			Article: newsReader.Article{Url: "https://tagesschau.de", Title: "Ein Testlauf ..."},
		},
		{
			Name: "pass with empty article",

			PublishFn: func(a newsReader.Article, eType string) error {
				if a.ID == "" {
					t.Fatalf("Article id must not be empty")
				}
				if eType != wantedEType {
					t.Fatalf("wanted eventType=%v got=%v", wantedEType, eType)
				}
				return nil
			},
			PublishInvoked: true,

			EType:   wantedEType,
			wantErr: false,
			Article: newsReader.Article{},
		},
		{
			Name: "queue error",

			PublishFn: func(a newsReader.Article, eType string) error {
				return fmt.Errorf("some queue error")
			},
			PublishInvoked: true,

			EType:   wantedEType,
			wantErr: true,
			Article: newsReader.Article{Url: "https://tagesschau.de", Title: "Ein Testlauf ..."},
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
			test.Name, func(t *testing.T) {
				q := &mock.Queue{PublishFn: test.PublishFn}
				p := eventStore.NewPublisher(q, test.EType, logger)

				err := p.Publish(test.Article)

				if (err != nil) != test.wantErr {
					t.Fatalf("want Publish to return none nil error, got %v", err)
				}

				if test.PublishInvoked != q.PublishInvoked {
					t.Fatalf("want PublishInvoked=%v got %v", test.PublishInvoked, q.PublishInvoked)
				}
			},
		)
	}
}
