package eventStore

import (
	"go.uber.org/zap"
	"newsReader"
)

type Consumer struct {
	queue newsReader.Queue
	eType string
	log   *zap.SugaredLogger
}

func NewConsumer(q newsReader.Queue, eType string, l *zap.SugaredLogger) *Consumer {
	return &Consumer{
		queue: q,
		eType: eType,
		log:   l,
	}
}
func (c Consumer) Consume(a chan<- newsReader.Article) {
	c.queue.Consume(c.eType, a)
}
