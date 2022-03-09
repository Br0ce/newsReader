package newsReader_test

import (
	"fmt"
	"testing"

	"go.uber.org/zap"
	"newsReader"
	"newsReader/mock"
)

func TestCollectorCollect(t *testing.T) {
	tests := []struct {
		Name string

		CrawlFn      func() ([]newsReader.Article, error)
		CrawlInvoked bool

		PublishFn      func(a newsReader.Article) error
		PublishInvoked bool

		wantErr bool
	}{
		{
			Name: "pass",

			CrawlFn: func() ([]newsReader.Article, error) {
				return []newsReader.Article{{}, {}}, nil
			},
			CrawlInvoked: true,

			PublishFn: func(a newsReader.Article) error {
				return nil
			},
			PublishInvoked: true,

			wantErr: false,
		},
		{
			Name: "Crawler error",

			CrawlFn: func() ([]newsReader.Article, error) {
				return nil, fmt.Errorf("some Crawler error")
			},
			CrawlInvoked: true,

			PublishFn: func(a newsReader.Article) error {
				return nil
			},
			PublishInvoked: false,

			wantErr: true,
		},
		{
			Name: "publisher error",

			CrawlFn: func() ([]newsReader.Article, error) {
				return []newsReader.Article{{Title: "test"}}, nil
			},
			CrawlInvoked: true,

			PublishFn: func(a newsReader.Article) error {
				return fmt.Errorf("some Publisher error")
			},
			PublishInvoked: true,

			wantErr: true,
		},
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	zapper, err := cfg.Build()
	if err != nil {
		t.Fatalf("could not init logger")
	}

	logger := zapper.Sugar()

	const maxWorker = 10

	for _, test := range tests {

		t.Run(
			test.Name, func(t *testing.T) {

				c := &mock.Crawler{CrawlFn: test.CrawlFn}
				p := &mock.Publisher{PublishFn: test.PublishFn}

				for n := 1; n < maxWorker; n++ {

					cb := newsReader.NewCollectorBuilder()
					clr, err := cb.Crawlers(c).Publisher(p).NumWorker(n).Logger(logger).Build()
					if err != nil {
						t.Fatalf("could not get new collector")
						return
					}

					err = clr.RunOnce()

					if (err != nil) != test.wantErr {
						t.Fatalf("want collect to return none nil error, got %v", err)
					}

					if test.CrawlInvoked != c.CrawlInvoked {
						t.Fatalf("want CrawlInvoked=%v got %v", test.CrawlInvoked, c.CrawlInvoked)
					}

					if test.PublishInvoked != p.PublishInvoked {
						t.Fatalf("want PublishInvoked=%v got %v", test.PublishInvoked, p.PublishInvoked)
					}
				}
			},
		)
	}
}
