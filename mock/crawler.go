package mock

import "newsReader"

type Crawler struct {
	CrawlFn      func() ([]newsReader.Article, error)
	CrawlInvoked bool
}

func (c *Crawler) Name() string {
	return "mockCrawler"
}

func (c *Crawler) Crawl() ([]newsReader.Article, error) {
	c.CrawlInvoked = true
	return c.CrawlFn()
}
