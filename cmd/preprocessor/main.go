package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"newsReader"
	"newsReader/eventStore"
	"newsReader/tsClient"
)

func main() {
	debug := flag.Bool("debug", false, "set loglevel to debug")
	env := flag.String("env-file", "./conf/.env", "set path to env-file")
	flag.Parse()

	cfg := zap.NewProductionConfig()
	if *debug {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
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
	esAddr, ok := os.LookupEnv("ES_ADDR")
	if !ok {
		log.Fatal("could not read eventstore addr from .env")
	}

	queue, err := eventStore.NewQueue(usr, pwd, esAddr, log.Named("queue"))
	if err != nil {
		log.Fatalf("could not create eventStore, %v\n", err.Error())
	}

	tsAddr, ok := os.LookupEnv("TS_ADDR")
	if !ok {
		log.Fatal("could not read torchServe addr from .env")
	}

	summary, err := tsClient.NewSummary(tsAddr, log.Named("summary"), time.Minute*2)
	if err != nil {
		log.Fatalf("could not init summary, %v\n", err.Error())
	}
	ner, err := tsClient.NewNER(tsAddr, log.Named("ner"), time.Second*30)
	if err != nil {
		log.Fatalf("could not init summary, %v\n", err.Error())
	}

	con := eventStore.NewConsumer(queue, "collected", log.Named("consumer-collected"))
	pub := eventStore.NewPublisher(queue, "preprocessed", log.Named("publisher-preprocessed"))

	ob := newsReader.NewOperatorBuilder()
	preprocessor, err := ob.Consumer(con).
		Publisher(pub).
		NumWorker(2).
		Processors(summary, ner).
		Logger(log.Named("operator")).
		Build()
	if err != nil {
		log.Fatalf("could not build preprocessor, %v\n", err)
	}

	err = preprocessor.Run()

	if err != nil {
		log.Fatalf("preprocessor finished with error, %v\n", err)
		os.Exit(1)
	}

}
