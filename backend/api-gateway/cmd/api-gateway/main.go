package main

import (
	"context"
	"flag"

	"github.com/hyperxpizza/eda/backend/api-gateway/internal/api"
	"github.com/hyperxpizza/eda/backend/utils"
	"github.com/sirupsen/logrus"
)

var producerConfigOpt = flag.String("producerConfig", "", "path to .properties file which contains kafka producer config")
var portOpt = flag.String("port", "8888", "port")

func main() {

	flag.Parse()

	lgr := logrus.New()
	ctx, cancel := context.WithCancel(context.Background())

	producerConfig, err := utils.ReadConfig(*producerConfigOpt)
	if err != nil {
		lgr.Fatal(err)
	}

	producer, err := api.NewProducer(ctx, cancel, lgr, []string{"messages"}, producerConfig)
	if err != nil {
		lgr.Fatal(err)
	}

	go producer.Run()

	srv := api.NewServer(ctx, cancel, lgr).WithProducer(producer)
	srv.Run(*portOpt)
}

// go run main.go --producerConfig=/home/hyperxpizza/dev/golang/eda/backend/producer.properties
