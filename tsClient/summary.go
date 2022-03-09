package tsClient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"newsReader"
)

type Summary struct {
	url     *url.URL
	timeout time.Duration
	log     *zap.SugaredLogger
}

func NewSummary(addr string, l *zap.SugaredLogger, timeout time.Duration) (*Summary, error) {
	u, err := url.Parse(fmt.Sprintf("http://%v/predictions/summarization", addr))
	if err != nil {
		return nil, err
	}

	return &Summary{url: u, log: l, timeout: timeout}, nil
}

func (s Summary) Name() string {
	return "Summary"
}

func (s Summary) Process(a newsReader.Article) (newsReader.Article, error) {
	s.log.Infow("summarize article", "method", "Process", "articleID", a.ID)
	bytes, err := post(s.url, strings.NewReader(a.Body), s.timeout)
	if err != nil {
		return newsReader.Article{}, err
	}

	type res struct {
		Summary string `json:"summary"`
	}
	var r res
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return newsReader.Article{}, err
	}

	a.Summary = r.Summary
	return a, nil
}
