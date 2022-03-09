package eventStore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"go.uber.org/zap"
	"newsReader"
)

type Queue struct {
	db        *esdb.Client
	log       *zap.SugaredLogger
	timeout   time.Duration
	batchSize uint64
}

func NewQueue(user, pwd, addr string, log *zap.SugaredLogger) (*Queue, error) {
	conn, err := esdb.ParseConnectionString(fmt.Sprintf("esdb://%s:%s@%s?tls=false", user, pwd, addr))
	if err != nil {
		return nil, fmt.Errorf("could not parse connection string, %w", err)
	}

	db, err := esdb.NewClient(conn)
	if err != nil {
		return nil, fmt.Errorf("could not create new client, %w", err)
	}

	return &Queue{
		db:        db,
		log:       log,
		batchSize: 30,
		timeout:   time.Second * 10,
	}, nil
}

func (q Queue) Publish(a newsReader.Article, eType string) error {
	q.log.Debugw("publish article", "method", "Publish", "articleID", a.ID, "eventType", eType)
	bytes, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("could not marshal articleID=%v, %w", a.ID, err)
	}

	event := esdb.EventData{
		ContentType: esdb.JsonContentType,
		EventType:   eType,
		Data:        bytes,
	}

	ctx, cancel := context.WithTimeout(context.Background(), q.timeout)
	defer cancel()
	_, err = q.db.AppendToStream(ctx, a.ID, esdb.AppendToStreamOptions{}, event)
	if err != nil {
		return fmt.Errorf("could not append eventType=%v to streamID=%v, %w", eType, a.ID, err)
	}

	return nil
}

func (q Queue) Consume(eType string, c chan<- newsReader.Article) {
	q.log.Debugw("consume", "method", "Consume", "eventType", eType)

	stream, err := q.db.SubscribeToStream(
		context.Background(), fmt.Sprintf("$et-%s", eType), esdb.SubscribeToStreamOptions{ResolveLinkTos: true},
	)
	if err != nil {
		q.log.Errorw("could not subscribe to stream", "method", "Consume", "errMsg", err)
		close(c)
		return
	}
	defer func(stream *esdb.Subscription) {
		err = stream.Close()
		if err != nil {
			q.log.Errorw("could not close stream", "method", "Consume", "errMsg", err)
		}
	}(stream)

	q.loopStream(stream, c)
}

func (q Queue) loopStream(stream *esdb.Subscription, c chan<- newsReader.Article) {
	for {
		evt := stream.Recv()

		var a newsReader.Article
		err := json.Unmarshal(evt.EventAppeared.Event.Data, &a)
		if err != nil {
			q.log.Errorw("could not unmarshal article", "method", "loopStream", "errMsg", err)
			close(c)
			return
		}

		q.log.Debugw(
			"append to articles",
			"method", "loopStream",
			"articleID", a.ID,
			"url", a.Url,
			"title", a.Title,
		)
		c <- a
	}
}
