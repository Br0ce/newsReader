package newsReader

import (
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Operator struct {
	processors []Processor
	con        Consumer
	pub        Publisher
	log        *zap.SugaredLogger
	numWorker  int
	tasks      chan Article
}

func NewOperatorBuilder() *OperatorBuilder {
	return &OperatorBuilder{}
}

type OperatorBuilder struct {
	pp []Processor
	c  Consumer
	p  Publisher
	l  *zap.SugaredLogger
	n  int
}

func (b *OperatorBuilder) Processors(pp ...Processor) *OperatorBuilder {
	b.pp = pp
	return b
}

func (b *OperatorBuilder) Consumer(c Consumer) *OperatorBuilder {
	b.c = c
	return b
}

func (b *OperatorBuilder) Publisher(p Publisher) *OperatorBuilder {
	b.p = p
	return b
}

func (b *OperatorBuilder) Logger(l *zap.SugaredLogger) *OperatorBuilder {
	b.l = l
	return b
}

func (b *OperatorBuilder) NumWorker(n int) *OperatorBuilder {
	b.n = n
	return b
}

func (b *OperatorBuilder) Build() (*Operator, error) {
	if b.l == nil {
		return nil, errors.New("no logger provided")
	}
	if b.c == nil {
		return nil, errors.New("no consumer provided")
	}
	if b.p == nil {
		return nil, errors.New("no pub provided")
	}
	if b.n < 1 {
		b.n = 1
		b.l.Warnw("numWorker < 1, set to 1", "method", "Build")
	}
	if len(b.pp) == 0 {
		b.pp = []Processor{}
	}

	return &Operator{
		log:        b.l,
		con:        b.c,
		pub:        b.p,
		numWorker:  b.n,
		processors: b.pp,
	}, nil
}

func (opr Operator) Run() error {
	opr.tasks = make(chan Article, opr.numWorker)

	opr.log.Infow("setup worker pool", "method", "Run", "numWorker", strconv.Itoa(opr.numWorker))
	eg := new(errgroup.Group)
	for i := 0; i < opr.numWorker; i++ {
		opr.log.Debugw("start operating", "method", "Run", "id", i)
		eg.Go(opr.operate)
	}

	go opr.con.Consume(opr.tasks)

	return eg.Wait()
}

func (opr Operator) operate() error {
	var ee []error

	for a := range opr.tasks {
		opr.log.Debugw("received article", "method", "operate", "articleID", a.ID)

		a, errs := opr.preprocess(a)
		if len(errs) != 0 {
			ee = append(ee, errs...)
		}

		err := opr.pub.Publish(a)
		if err != nil {
			opr.log.Warnw(
				"publish error",
				"method", "operate",
				"articleID", a.ID,
				"errMsg", err.Error(),
			)
			ee = append(ee, fmt.Errorf("publish article with ID=%v failed, %w", a.ID, err))
			continue
		}
	}

	if len(ee) != 0 {
		retError := errors.New("operate error")
		for _, e := range ee {
			retError = fmt.Errorf("%v, %w", e.Error(), retError)
		}
		return retError
	}
	return nil
}

func (opr Operator) preprocess(a Article) (Article, []error) {
	opr.log.Debugw("preprocess article", "method", "preprocess", "articleID", a.ID)

	var ee []error
	for _, p := range opr.processors {

		tmp, err := p.Process(a)
		if err != nil {
			opr.log.Warnw(
				"process error",
				"method", "operate",
				"articleID", a.ID,
				"processorName", p.Name(),
				"errMsg", err.Error(),
			)

			ee = append(ee, fmt.Errorf("processor=%s on article with id=%s failed, %w", p.Name(), a.ID, err))
			continue
		}

		a = tmp
	}

	return a, ee
}
