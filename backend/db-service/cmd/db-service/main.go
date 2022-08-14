package main

import (
	"context"
	"flag"

	"github.com/hyperxpizza/eda/backend/db-service/internal/config"
	"github.com/hyperxpizza/eda/backend/db-service/internal/consumer"
	"github.com/hyperxpizza/eda/backend/db-service/internal/database"
	"github.com/hyperxpizza/eda/backend/utils"
	"github.com/sirupsen/logrus"
)

//go run main.go --postgresConfig="/home/hyperxpizza/dev/golang/eda/backend/postgresConfig.json" --consumerConfig="/home/hyperxpizza/dev/golang/eda/backend/consumer.properties"
var postgresConfigOpt = flag.String("postgresConfig", "", "path to .json file containing postgres config")
var consumerConfigOpt = flag.String("consumerConfig", "", "path to .properties file containing kafka consumer config")

func main() {
	flag.Parse()
	lgr := logrus.New()

	postrgesConfig, err := config.NewConfig(*postgresConfigOpt)
	if err != nil {
		lgr.Fatal(err)
	}

	consumerConfig, err := utils.ReadConfig(*consumerConfigOpt)
	if err != nil {
		lgr.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	pgClient, err := database.NewPostgresClient(postrgesConfig, lgr)
	if err != nil {
		lgr.Fatal(err)
	}

	consumer, err := consumer.NewConsumer(ctx, cancel, lgr, pgClient, []string{"messages"}, consumerConfig)
	if err != nil {
		lgr.Fatal(err)
	}

	consumer.Run()

}
