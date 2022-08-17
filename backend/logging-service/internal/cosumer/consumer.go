package consumer

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	ctx      context.Context
	cancel   context.CancelFunc
	lgr      logrus.FieldLogger
	consumer *kafka.Consumer
	wg       sync.WaitGroup
	writers  map[io.Writer]os.File
}

func NewConsumer(ctx context.Context, cancel context.CancelFunc, lgr logrus.FieldLogger, cfg kafka.ConfigMap) {

}

func (c *Consumer) Run() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			event := c.consumer.Poll(100)
			if event == nil {
				continue
			}
		}
	}
}

func (c *Consumer) handleMessage()
