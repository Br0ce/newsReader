package newsReader

import (
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Collector struct {
	crawlers  []Crawler
	pub       Publisher
	log       *zap.SugaredLogger
	tasks     chan Crawler
	numWorker int
}

func NewCollectorBuilder() *CollectorBuilder {
	return &CollectorBuilder{}
}

type CollectorBuilder struct {
	cc []Crawler
	p  Publisher
	l  *zap.SugaredLogger
	n  int
}

func (b *CollectorBuilder) Crawlers(cc ...Crawler) *CollectorBuilder {
	b.cc = cc
	return b
}

func (b *CollectorBuilder) Publisher(p Publisher) *CollectorBuilder {
	b.p = p
	return b
}

func (b *CollectorBuilder) Logger(l *zap.SugaredLogger) *CollectorBuilder {
	b.l = l
	return b
}

func (b *CollectorBuilder) NumWorker(n int) *CollectorBuilder {
	b.n = n
	return b
}

func (b *CollectorBuilder) Build() (*Collector, error) {
	if b.l == nil {
		return nil, errors.New("no logger provided")
	}
	if b.p == nil {
		return nil, errors.New("no publisher provided")
	}
	if b.n < 1 {
		b.n = 1
		b.l.Warnw("numWorker < 1, set to 1", "method", "Build")
	}
	if len(b.cc) == 0 {
		b.cc = []Crawler{}
	}

	return &Collector{
		crawlers:  b.cc,
		pub:       b.p,
		log:       b.l,
		numWorker: b.n,
	}, nil
}

func (clr Collector) RunOnce() error {
	clr.tasks = make(chan Crawler, clr.numWorker)

	clr.log.Infow("setup worker pool", "method", "RunOnce", "numWorker", clr.numWorker)
	eg := new(errgroup.Group)
	for i := 0; i < clr.numWorker; i++ {
		clr.log.Debugw("start collecting", "method", "RunOnce", "id", i)
		eg.Go(clr.collect)
	}

	clr.log.Infow(
		"start collecting",
		"method", "RunOnce",
		"numCrawler", strconv.Itoa(len(clr.crawlers)),
	)

	for i, c := range clr.crawlers {
		clr.log.Debugw("add task", "method", "RunOnce", "id", i)
		clr.tasks <- c
	}

	close(clr.tasks)
	return eg.Wait()
}

func (clr Collector) collect() error {
	for c := range clr.tasks {
		clr.log.Debugw("crawling resource", "method", "collect", "resource", c.Name())

		articles, err := c.Crawl()
		if err != nil {
			return fmt.Errorf("could not crawl resource=%s, %w", c.Name(), err)
		}

		for _, a := range articles {
			clr.log.Debugw("publish articles", "method", "collect", "title", a.Title)

			err = clr.pub.Publish(a)
			if err != nil {
				return fmt.Errorf("could not publish article with id=%s, %w", a.ID, err)
			}
		}

	}
	clr.log.Debugw("finished collecting", "method", "collect")
	return nil
}
