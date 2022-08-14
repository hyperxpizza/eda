package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

type Producer struct {
	producer     *kafka.Producer
	lgr          logrus.FieldLogger
	ctx          context.Context
	cancel       context.CancelFunc
	topics       map[string]struct{}
	deliveryChan chan kafka.Event
}

func NewProducer(ctx context.Context, cancel context.CancelFunc, lgr logrus.FieldLogger, topics []string, cfg kafka.ConfigMap) (*Producer, error) {
	p, err := kafka.NewProducer(&cfg)
	if err != nil {
		return nil, err
	}

	producer := &Producer{
		producer: p,
		lgr:      lgr.WithField("module", "kafka-producer"),
		ctx:      ctx,
		topics:   make(map[string]struct{}, 0),
	}

	if err := producer.CreateTopics(topics); err != nil {
		return nil, err
	}

	return producer, nil
}

func (p *Producer) CreateTopics(names []string) error {

	admin, err := kafka.NewAdminClientFromProducer(p.producer)
	if err != nil {
		p.lgr.Errorf("could not create admin client: %s", err.Error())
		return err
	}

	topics := make([]kafka.TopicSpecification, 0)

	for _, name := range names {
		topic := kafka.TopicSpecification{Topic: name, NumPartitions: 1}
		topics = append(topics, topic)
	}

	res, err := admin.CreateTopics(p.ctx, topics)
	if err != nil {
		p.lgr.Errorf("could not create topics%v: %s", topics, err.Error())
		return err
	}

	for _, topic := range res {
		if topic.Error.Code() == kafka.ErrNoError {
			if _, ok := p.topics[topic.Topic]; !ok {
				p.topics[topic.Topic] = struct{}{}
			}
		} else {
			p.lgr.Errorf("could not create a topic(%s) error: %s", topic.Topic, topic.Error.String())
		}
	}

	if len(p.topics) == 0 {
		p.lgr.Info("did not not create any topics")
	}

	return nil
}

func (p *Producer) Run() {
	p.lgr.Info("running kafka-producer...")
	for {
		select {
		case <-p.ctx.Done():
			p.producer.Close()
			close(p.deliveryChan)
			return
		case e := <-p.deliveryChan:
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					p.lgr.Errorf("delivery failed: %s", ev.TopicPartition.Error.Error())
				} else {
					p.lgr.Infof("delivered a message to topic(%s) [%d] at offset %v", *ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			default:
				p.lgr.Debugf("[deliveryChan] ignored event: %s", ev)

			}
		case e := <-p.producer.Events():
			switch ev := e.(type) {
			case kafka.Error:
				p.lgr.Errorf("error: %s", ev.Error())
			default:
				p.lgr.Debugf("ignored event: %s", ev)
			}
		}
	}

}

func (p *Producer) WriteMessage(topic string, value interface{}) error {

	valB, err := json.Marshal(value)
	if err != nil {
		return err
	}

	topicPartition := kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}
	msg := kafka.Message{
		TopicPartition: topicPartition,
		Value:          valB,
		Timestamp:      time.Now(),
		TimestampType:  kafka.TimestampCreateTime,
	}

	return p.producer.Produce(&msg, p.deliveryChan)
}
