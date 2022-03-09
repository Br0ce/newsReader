package newsReader

import (
	"crypto/md5"
	"fmt"
	"net/url"
)

type Article struct {
	ID        string   `json:"id"`
	Author    string   `json:"author"`
	Body      string   `json:"body"`
	Title     string   `json:"title"`
	Created   string   `json:"created"`
	Collected string   `json:"collected"`
	Url       string   `json:"url"`
	Summary   string   `json:"summary"`
	Tags      []string `json:"tags"`
	Pers      []string `json:"pers"`
	Locs      []string `json:"locs"`
	Orgs      []string `json:"orgs"`
}

func ArticleID(a Article) string {
	u, err := url.Parse(a.Url)
	if err != nil {
		// md5.Sum("") == d41d8cd98f00b204e9800998ecf8427e
		return fmt.Sprintf("article-%v", id(a.Title))
	}

	return fmt.Sprintf("article-%v", id(u.Host+a.Title))
}

func id(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
