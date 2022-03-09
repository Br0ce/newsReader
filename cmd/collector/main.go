package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"newsReader"
	"newsReader/colly"
	"newsReader/eventStore"
)

func main() {
	debug := flag.Bool("debug", false, "set loglevel to debug")
	env := flag.String("env-file", "./conf/.env", "set path to env-file")
	flag.Parse()

	cfg := zap.NewProductionConfig()
	if *debug {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	zapper, err := cfg.Build()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could init logger, %v\n", err.Error())
		os.Exit(1)
	}

	log := zapper.Sugar()

	err = godotenv.Load(*env)
	if err != nil {
		log.Fatalf("could load env-file=%s, %v\n", *env, err.Error())
	}

	usr, ok := os.LookupEnv("ES_USER")
	if !ok {
		log.Fatal("could not read eventstore user from .env")
	}
	pwd, ok := os.LookupEnv("ES_PWD")
	if !ok {
		log.Fatal("could not read eventstore pwd from .env")
	}
	addr, ok := os.LookupEnv("ES_ADDR")
	if !ok {
		log.Fatal("could not read eventstore addr from .env")
	}

	queue, err := eventStore.NewQueue(usr, pwd, addr, log.Named("queue"))
	if err != nil {
		log.Fatalf("could not create eventStore, %v\n", err.Error())
	}

	c := colly.NewTagesschauCrawler(log.Named("tagesschau"))
	p := eventStore.NewPublisher(queue, "collected", log.Named("publisher-collected"))

	cb := newsReader.NewCollectorBuilder()
	collector, err := cb.Crawlers(c).Publisher(p).NumWorker(1).Logger(log.Named("collector")).Build()
	if err != nil {
		log.Fatalf("could not build collector, %v\n", err.Error())
	}

	ticker := time.NewTicker(time.Hour * 2)
	stop := time.After(time.Hour * 48)

	for {
		// start immediately
		err = collector.RunOnce()
		if err != nil {
			log.Fatalf("collector finished with error, %v\n", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-stop:
			os.Exit(0)
		}
	}

}
