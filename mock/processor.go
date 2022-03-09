package mock

import "newsReader"

type Processor struct {
	ProcessFn      func(a newsReader.Article) (newsReader.Article, error)
	ProcessInvoked bool
}

func (p *Processor) Name() string {
	return "mockProcessor"
}

func (p *Processor) Process(a newsReader.Article) (newsReader.Article, error) {
	p.ProcessInvoked = true
	return p.ProcessFn(a)
}
