package openSearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"go.uber.org/zap"
	"newsReader"
)

type Publisher struct {
	client *opensearch.Client
	log    *zap.SugaredLogger
}

func NewPublisher(user, pwd, addr string, l *zap.SugaredLogger) (*Publisher, error) {
	cfg := opensearch.Config{
		Addresses: []string{
			fmt.Sprintf("https://%s", addr),
		},
		Username: user,
		Password: pwd,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second * 20,
			DialContext:           (&net.Dialer{Timeout: time.Second * 20}).DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	ping, err := client.Ping()
	if err != nil {
		return nil, err
	}
	if ping.StatusCode >= 400 {
		return nil, fmt.Errorf("could not ping opensearch, status code=%v", ping.StatusCode)
	}

	return &Publisher{client: client, log: l}, nil
}

func (p Publisher) Publish(a newsReader.Article) error {
	p.log.Infow("publish article", "method", "Publish", "articleID", a.ID)
	b, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("could not marshal article with id=%s, %w", a.ID, err)
	}

	request := opensearchapi.IndexRequest{Index: "article-1", DocumentID: a.ID, Body: bytes.NewReader(b)}
	resp, err := request.Do(context.Background(), p.client)
	if err != nil {
		return fmt.Errorf("could not request publish index request article with id=%s, %w", a.ID, err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf(
			"opensearch response status code=%v while publishing article with id=%s", resp.StatusCode, a.ID,
		)
	}
	return nil
}
