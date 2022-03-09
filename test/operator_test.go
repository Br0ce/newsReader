package newsReader_test

import (
	"errors"
	"testing"

	"go.uber.org/zap"
	"newsReader"
	"newsReader/mock"
)

func TestOperatorRunOnce(t *testing.T) {

	tests := []struct {
		Name string

		ConsumeFn      func(c chan<- newsReader.Article)
		ConsumeInvoked bool

		ProcessFn      func(a newsReader.Article) (newsReader.Article, error)
		ProcessInvoked bool

		PublishFn      func(a newsReader.Article) error
		PublishInvoked bool

		wantErr bool
	}{
		{
			Name: "pass",

			ConsumeFn: func(c chan<- newsReader.Article) {
				c <- newsReader.Article{Title: "aa"}
				c <- newsReader.Article{Title: "bb"}
				c <- newsReader.Article{Title: "cc"}
				close(c)
			},
			ConsumeInvoked: true,

			PublishFn: func(a newsReader.Article) error {
				return nil
			},
			PublishInvoked: true,

			ProcessFn: func(a newsReader.Article) (newsReader.Article, error) {
				if a.Title == "aa" {
					a.ID = "00"
					return a, nil
				}
				if a.Title == "bb" {
					a.ID = "11"
					return a, nil
				}
				if a.Title == "cc" {
					a.ID = "22"
					return a, nil
				}
				t.Fatalf("articel unknown")
				return newsReader.Article{}, nil
			},
			ProcessInvoked: true,

			wantErr: false,
		},
		{
			Name: "consumer closed",

			ConsumeFn: func(c chan<- newsReader.Article) {
				close(c)
			},
			ConsumeInvoked: true,

			PublishFn: func(a newsReader.Article) error {
				return nil
			},
			PublishInvoked: false,

			ProcessFn: func(a newsReader.Article) (newsReader.Article, error) {
				return newsReader.Article{}, nil
			},
			ProcessInvoked: false,

			wantErr: false,
		},
		{
			Name: "process error",

			ConsumeFn: func(c chan<- newsReader.Article) {
				c <- newsReader.Article{Title: "aa"}
				c <- newsReader.Article{Title: "bb"}
				close(c)
			},
			ConsumeInvoked: true,

			ProcessFn: func(a newsReader.Article) (newsReader.Article, error) {
				return newsReader.Article{}, errors.New("some process error")
			},
			ProcessInvoked: true,

			PublishFn: func(a newsReader.Article) error {
				return nil
			},
			PublishInvoked: true,

			wantErr: true,
		},
		{
			Name: "publisher error",
			ConsumeFn: func(c chan<- newsReader.Article) {
				c <- newsReader.Article{Title: "aa"}
				c <- newsReader.Article{Title: "bb"}
				close(c)
			},
			ConsumeInvoked: true,

			ProcessFn: func(a newsReader.Article) (newsReader.Article, error) {
				if a.Title == "aa" {
					a.ID = "00"
					return a, nil
				}
				if a.Title == "bb" {
					a.ID = "11"
					return a, nil
				}
				t.Fatalf("articel unknown")
				return newsReader.Article{}, nil
			},
			ProcessInvoked: true,

			PublishFn: func(a newsReader.Article) error {
				return errors.New("some publisher error")
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

				c := &mock.Consumer{ConsumeFn: test.ConsumeFn}
				pu := &mock.Publisher{PublishFn: test.PublishFn}
				pr := &mock.Processor{ProcessFn: test.ProcessFn}

				for n := 1; n < maxWorker; n++ {

					oprB := newsReader.NewOperatorBuilder()
					opr, err := oprB.Processors(pr).Publisher(pu).Consumer(c).Logger(logger).Build()

					if err != nil {
						t.Fatalf("could not get new preprocessor")
						return
					}

					err = opr.Run()

					if (err != nil) != test.wantErr {
						t.Fatalf("wanted return error=%v got=%v", test.wantErr, err.Error())
					}
					if c.ConsumeInvoked != test.ConsumeInvoked {
						t.Fatalf("want ConsumeInvoked=%v got=%v", test.ConsumeInvoked, c.ConsumeInvoked)
					}
					if pu.PublishInvoked != test.PublishInvoked {
						t.Fatalf("want PublishInvoked=%v got=%v", test.PublishInvoked, pu.PublishInvoked)
					}
					if pr.ProcessInvoked != test.ProcessInvoked {
						t.Fatalf("want ProcessInvoked=%v got=%v", test.ProcessInvoked, pr.ProcessInvoked)
					}
				}
			},
		)
	}
}
