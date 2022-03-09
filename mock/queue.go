package mock

import "newsReader"

type Publisher struct {
	PublishFn      func(a newsReader.Article) error
	PublishInvoked bool
}

func (p *Publisher) Publish(a newsReader.Article) error {
	p.PublishInvoked = true
	return p.PublishFn(a)
}

type Consumer struct {
	ConsumeFn      func(c chan<- newsReader.Article)
	ConsumeInvoked bool
}

func (co *Consumer) Consume(c chan<- newsReader.Article) {
	co.ConsumeInvoked = true
	co.ConsumeFn(c)
}

type Queue struct {
	PublishFn      func(a newsReader.Article, eType string) error
	PublishInvoked bool

	ConsumeFn      func(eType string, c chan<- newsReader.Article)
	ConsumeInvoked bool
}

func (q *Queue) Publish(a newsReader.Article, eType string) error {
	q.PublishInvoked = true
	return q.PublishFn(a, eType)
}

func (q *Queue) Consume(eType string, c chan<- newsReader.Article) {
	q.ConsumeInvoked = true
	q.ConsumeFn(eType, c)
}
