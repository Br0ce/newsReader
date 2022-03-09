package eventStore

import (
	"go.uber.org/zap"
	"newsReader"
)

type Publisher struct {
	queue newsReader.Queue
	eType string
	log   *zap.SugaredLogger
}

func NewPublisher(q newsReader.Queue, eType string, l *zap.SugaredLogger) *Publisher {
	return &Publisher{queue: q, eType: eType, log: l}
}

func (p Publisher) Publish(a newsReader.Article) error {
	if len(a.ID) == 0 {
		a.ID = newsReader.ArticleID(a)
		p.log.Debugw(
			"setting new articleID",
			"method", "Publish",
			"articleID", a.ID,
			"title", a.Title,
			"url", a.Url,
		)
	}

	p.log.Infow("publish article", "method", "Publish", "articleID", a.ID, "eventType", p.eType)
	return p.queue.Publish(a, p.eType)
}
