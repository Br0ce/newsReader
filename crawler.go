package newsReader

type Crawler interface {
	Name() string
	Crawl() ([]Article, error)
}
