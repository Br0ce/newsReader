package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"newsReader"
	"newsReader/eventStore"
	"newsReader/openSearch"
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

	esUser, ok := os.LookupEnv("ES_USER")
	if !ok {
		log.Fatal("could not read eventstore user from .env")
	}
	esPwd, ok := os.LookupEnv("ES_PWD")
	if !ok {
		log.Fatal("could not read eventstore pwd from .env")
	}
	esAddr, ok := os.LookupEnv("ES_ADDR")
	if !ok {
		log.Fatal("could not read eventstore addr from .env")
	}

	queue, err := eventStore.NewQueue(esUser, esPwd, esAddr, log.Named("queue"))
	if err != nil {
		log.Fatalf("could not create eventStore, %v\n", err.Error())
	}
	con := eventStore.NewConsumer(queue, "preprocessed", log.Named("consumer-preprocessed"))

	osUser, ok := os.LookupEnv("OS_USER")
	if !ok {
		log.Fatal("could not read opensearch user from .env")
	}
	osPwd, ok := os.LookupEnv("OS_PWD")
	if !ok {
		log.Fatal("could not read opensearch pwd from .env")
	}
	osAddr, ok := os.LookupEnv("OS_ADDR")
	if !ok {
		log.Fatal("could not read opensearch addr from .env")
	}

	pub, err := openSearch.NewPublisher(osUser, osPwd, osAddr, log.Named("publisher-openSearch"))
	if err != nil {
		log.Fatalf("could not create new openSearch publisher, %v\n", err.Error())
	}

	ab := newsReader.NewOperatorBuilder()
	archiver, err := ab.Consumer(con).
		Publisher(pub).
		NumWorker(2).
		Logger(log.Named("operator")).
		Build()
	if err != nil {
		log.Fatalf("could not build archiver, %v\n", err)
	}

	err = archiver.Run()

	if err != nil {
		log.Fatalf("archiver finished with error, %v\n", err)
		os.Exit(1)
	}

}
