package newsReader_test

import (
	"strings"
	"testing"

	"newsReader"
)

func TestArticleID(t *testing.T) {
	tests := []struct {
		name string
		arg  newsReader.Article
	}{
		{
			name: "https",
			arg:  newsReader.Article{Url: "https//www.google.com", Title: "some title"},
		},
		{
			name: "http",
			arg:  newsReader.Article{Url: "https//www.google.com", Title: "some title"},
		},
		{
			name: "empty title",
			arg:  newsReader.Article{Url: "https//www.google.com"},
		},
		{
			name: "empty url",
			arg:  newsReader.Article{Title: "some title"},
		},
		{
			name: "empty article",
			arg:  newsReader.Article{},
		},
		{
			name: "invalid url",
			arg:  newsReader.Article{Url: "https:::.google.com++++\\"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := newsReader.ArticleID(tt.arg)
				if !strings.HasPrefix(got, "article-") {
					t.Errorf("want prefex article, got=%v", got)
				}
				if len(got) != 40 {
					t.Errorf("want articleID len=40, got=%v", len(got))
				}
			},
		)
	}
}
