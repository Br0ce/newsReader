package tsClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"newsReader"
)

type NER struct {
	url     *url.URL
	log     *zap.SugaredLogger
	maxLen  int
	timeout time.Duration
}

type response struct {
	Token string `json:"token"`
	Pred  string `json:"pred"`
}

func NewNER(addr string, l *zap.SugaredLogger, timeout time.Duration) (*NER, error) {
	u, err := url.Parse(fmt.Sprintf("http://%s/predictions/ner", addr))
	if err != nil {
		return nil, err
	}

	return &NER{url: u, log: l, maxLen: 512, timeout: timeout}, nil
}

func (n NER) Name() string {
	return "NER"
}

func (n NER) Process(a newsReader.Article) (newsReader.Article, error) {
	n.log.Infow("NER for article", "method", "Process", "articleID", a.ID)

	b := a.Body
	if len(a.Body) > n.maxLen {
		n.log.Debugw(
			"truncating article body",
			"method", "Process",
			"articleID", a.ID,
			"lenBody", strconv.Itoa(len(b)),
			"maxLen", strconv.Itoa(n.maxLen),
		)
		b = a.Body[:n.maxLen]
	}

	bytes, err := post(n.url, strings.NewReader(b), n.timeout)
	if err != nil {
		return newsReader.Article{}, err
	}

	var rr []response
	err = json.Unmarshal(bytes, &rr)
	if err != nil {
		return newsReader.Article{}, fmt.Errorf(
			"could not unmarshal response from torchServe for articel id=%s", a.ID,
		)
	}

	a.Pers, a.Locs, a.Orgs = entities(rr)

	return a, nil
}

func entities(rr []response) (pers, locs, orgs []string) {
	pers = []string{}
	locs = []string{}
	orgs = []string{}

	for i, r := range rr {

		if r.Pred == "O" {
			continue
		}

		if strings.HasPrefix(r.Pred, "I") {
			continue
		}

		if r.Pred == "B-PER" {
			p, _ := span(rr[i:], "PER")
			pers = append(pers, p)
			continue
		}

		if r.Pred == "B-LOC" {
			l, _ := span(rr[i:], "LOC")
			locs = append(locs, l)
			continue
		}

		if r.Pred == "B-ORG" {
			o, _ := span(rr[i:], "ORG")
			orgs = append(orgs, o)
			continue
		}

	}

	pers = removeDuplicates(pers)
	locs = removeDuplicates(locs)
	orgs = removeDuplicates(orgs)

	return
}

func span(rr []response, target string) (string, error) {
	if len(rr) == 0 {
		return "", errors.New("len rr must be > 0")
	}

	if rr[0].Pred != fmt.Sprintf("B-%s", target) {
		return "", fmt.Errorf("rr does not start with B-%s prediction, got=%s", target, rr[0].Pred)
	}

	parts := []string{rr[0].Token}

	for i := 1; i < len(rr); i++ {
		if rr[i].Pred != fmt.Sprintf("I-%s", target) {
			break
		}
		parts = append(parts, rr[i].Token)
	}

	return strings.Join(parts, " "), nil
}

func removeDuplicates(ss []string) []string {
	set := make(map[string]bool)
	res := []string{}
	for _, s := range ss {
		if _, ok := set[s]; !ok {
			set[s] = true
			res = append(res, s)
		}
	}
	return res
}
