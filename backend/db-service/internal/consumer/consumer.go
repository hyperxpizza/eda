package consumer

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hyperxpizza/eda/backend/db-service/internal/database"
	"github.com/hyperxpizza/eda/backend/utils"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	ctx      context.Context
	cancel   context.CancelFunc
	lgr      logrus.FieldLogger
	consumer *kafka.Consumer
	pgClient *database.PostgresClient
	wg       sync.WaitGroup
}

func NewConsumer(ctx context.Context, cancel context.CancelFunc, lgr logrus.FieldLogger, pgClient *database.PostgresClient, topics []string, cfg kafka.ConfigMap) (*Consumer, error) {
	c, err := kafka.NewConsumer(&cfg)
	if err != nil {
		lgr.Errorf("could not create a new kafka cosumer instance: %s", err.Error())
		return nil, err
	}

	err = c.SubscribeTopics(topics, nil)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		ctx:      ctx,
		cancel:   cancel,
		lgr:      lgr.WithField("module", "kafka-consumer"),
		consumer: c,
		pgClient: pgClient,
		wg:       sync.WaitGroup{},
	}, nil
}

func (c *Consumer) Run() {
	c.lgr.Info("running consumer...")
	for {
		select {
		case <-c.ctx.Done():
			c.consumer.Close()
			return
		default:
			event := c.consumer.Poll(100)
			if event == nil {
				continue
			}

			switch e := event.(type) {
			case *kafka.Message:
				if e.TopicPartition.Error != nil {
					c.lgr.Errorf("delivery failed:")
					continue
				}
				c.logMessage(e)
				c.wg.Add(1)
				go c.handleMessage(e)
			case kafka.Error:
				c.lgr.Error("error: %s", e.Error())
				if e.Code() == kafka.ErrAllBrokersDown {
					return
				}
			}
		}
	}
}

func (c *Consumer) Close() {
	c.wg.Wait()
	c.cancel()
}

func (c *Consumer) logMessage(e *kafka.Message) {
	c.lgr.Info("---new message")
	c.lgr.Infof("topic: %s", *e.TopicPartition.Topic)
	c.lgr.Info("headers:")
	for i, h := range e.Headers {
		c.lgr.Infof("%d. key: %s value: %s", i+1, h.Key, string(h.Value))
	}
	c.lgr.Infof("key: %s", string(e.Key))
	c.lgr.Infof("value: %s", string(e.Value))
	c.lgr.Info("end message---")
}

func (c *Consumer) handleMessage(msg *kafka.Message) {
	defer c.wg.Done()

	_, err := c.consumer.StoreMessage(msg)
	if err != nil {
		c.lgr.Errorf("could not store message: %s", err.Error())
	}

	if msg.TopicPartition.Topic == nil {
		c.lgr.Debug("could not determine message topic, return")
		return
	}

	switch *msg.TopicPartition.Topic {
	case "messages":
		var message utils.Message
		err := json.Unmarshal(msg.Value, &message)
		if err != nil {
			c.lgr.Errorf("could not unmarshal msg.Value into utils.Message: %s", err.Error())
			return
		}
		err = c.pgClient.InsertMessage(message.Content)
		if err != nil {
			c.lgr.Errorf("could not insert message into the database: %s", err.Error())
		}
	default:
		c.lgr.Debugf("unknown topic: %s", *msg.TopicPartition.Topic)
	}
}
