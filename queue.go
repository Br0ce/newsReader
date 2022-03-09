package newsReader

type Queue interface {
	Publish(a Article, eType string) error
	Consume(eType string, c chan<- Article)
}

type Publisher interface {
	Publish(a Article) error
}

type Consumer interface {
	Consume(c chan<- Article)
}
