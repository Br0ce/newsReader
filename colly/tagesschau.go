package colly

import (
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
	"newsReader"
)

type Tagesschau struct {
	log *zap.SugaredLogger
}

func NewTagesschauCrawler(l *zap.SugaredLogger) *Tagesschau {
	return &Tagesschau{log: l}
}

func (t Tagesschau) Name() string {
	return "tagesschau"
}

func (t *Tagesschau) Crawl() ([]newsReader.Article, error) {
	url := "https://www.tagesschau.de"
	t.log.Infow("start crawling", "method", "Crawl", "url", url)

	cltr := colly.NewCollector()
	articles := make([]newsReader.Article, 0)

	cltr.OnHTML(
		"a.teaser__link", func(elem *colly.HTMLElement) {
			link := elem.Attr("href")
			err := elem.Request.Visit(link)
			if err != nil {
				t.log.Errorf("could not visit link=%v", link)
			}
		},
	)

	cltr.OnHTML(
		"article.container", func(elem *colly.HTMLElement) {
			dom := elem.DOM
			titel := dom.Find("span.seitenkopf__headline--text").Text()
			created := dom.Find("div.metatextline").Text()

			// assemble body
			body := strings.Builder{}
			dom.Find("p.m-ten.m-offset-one.l-eight.l-offset-two.textabsatz.columns.twelve").Each(
				func(i int, s *goquery.Selection) {
					body.WriteString(s.Text())
					body.WriteString(" ")
				},
			)

			// assemble tags
			tags := make([]string, 0)
			dom.Find("a.tag-btn.tag-btn--light-grey").Each(
				func(i int, s *goquery.Selection) {
					tags = append(tags, s.Text())
				},
			)

			a := newsReader.Article{
				Url:       elem.Request.Ctx.Get("url"),
				Collected: elem.Request.Ctx.Get("date"),
				Tags:      tags,
				Created:   t.cleanDate(created),
				Body:      t.cleanBody(body.String()),
				Title:     titel,
			}

			articles = append(articles, a)
		},
	)

	var numVisited uint32
	cltr.OnRequest(
		func(r *colly.Request) {
			r.Ctx.Put("url", r.URL.String())
			r.Ctx.Put("date", time.Now().String())
			t.log.Debugw("visiting website", "method", "Crawl", "url", r.URL)
			atomic.AddUint32(&numVisited, 1)
		},
	)

	err := cltr.Visit(url)
	if err != nil {
		return nil, err
	}

	t.log.Infow(
		"finished crawling",
		"method", "Crawl",
		"numVisited", numVisited,
		"numArticles", strconv.Itoa(len(articles)),
	)
	return articles, nil
}

func (t Tagesschau) cleanBody(s string) string {
	return strings.ReplaceAll(s, "\n", "")
}

// todo cast to time
func (t Tagesschau) cleanDate(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.TrimLeft(s, " ")
	s = strings.TrimRight(s, " ")
	return s
}
