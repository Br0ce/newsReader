package newsReader

type Processor interface {
	Name() string
	Process(a Article) (Article, error)
}
