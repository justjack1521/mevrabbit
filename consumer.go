package mevrabbit

import (
	"context"
	"github.com/justjack1521/mevrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
)

type ConsumerContext struct {
	context.Context
	userID   uuid.UUID
	playerID uuid.UUID
	Delivery rabbitmq.Delivery
}

func (c *ConsumerContext) UserID() uuid.UUID {
	return c.userID
}

func (c *ConsumerContext) PlayerID() uuid.UUID {
	return c.playerID
}

type ConsumerHandler func(ctx *ConsumerContext) (rabbitmq.Action, error)

type StandardConsumer struct {
	actual     *rabbitmq.Consumer
	Queue      Queue
	RoutingKey RoutingKey
	Exchange   Exchange
	handler    ConsumerHandler
	closed     bool
}

func (s *StandardConsumer) Run() error {
	return s.actual.Run(s.standardConsumption())
}

func NewStandardConsumer(conn *rabbitmq.Conn, queue Queue, key RoutingKey, exchange Exchange, handler ConsumerHandler, opts ...func(*rabbitmq.ConsumerOptions)) (*StandardConsumer, error) {

	consumer := &StandardConsumer{
		Queue:      queue,
		RoutingKey: key,
		Exchange:   exchange,
		handler:    handler,
	}

	var options = []func(*rabbitmq.ConsumerOptions){
		rabbitmq.WithConsumerOptionsRoutingKey(string(key)),
		rabbitmq.WithConsumerOptionsExchangeName(string(exchange)),
	}

	opts = append(opts, options...)

	actual, err := rabbitmq.NewConsumer(conn, string(queue), opts...)
	if err != nil {
		return nil, err
	}

	consumer.actual = actual
	return consumer, nil
}

func (s *StandardConsumer) standardConsumption() rabbitmq.Handler {
	return func(d rabbitmq.Delivery) rabbitmq.Action {

		user, err := ExtractUserID(d)
		if err != nil {
			return rabbitmq.NackDiscard
		}

		player, err := ExtractPlayerID(d)
		if err != nil {
			return rabbitmq.NackDiscard
		}

		ctx := &ConsumerContext{
			Context:  mevrpc.NewOutgoingContext(context.Background(), user, player),
			Delivery: d,
			userID:   user,
			playerID: player,
		}

		result, _ := s.handler(ctx)
		return result
	}
}

func (s *StandardConsumer) Close() {
	if s.closed == true {
		return
	}
	s.closed = true
	s.actual.Close()
}

func (s *StandardConsumer) WithNewRelic(relic *newrelic.Application) {
	s.handler = consumerNewRelicMiddleWare(relic, s.handler)
}

func (s *StandardConsumer) WithLogging(logger *logrus.Logger) {
	s.handler = consumeLoggerMiddleWare(logger, s.handler)
}

func ExtractUserID(d rabbitmq.Delivery) (uuid.UUID, error) {
	id, ok := d.Headers[userIDHeaderKey]
	if ok == false {
		return uuid.Nil, errExtractClientIDFromMessageHeader(errClientIDNotFound)
	}
	client, err := uuid.FromString(id.(string))
	if err != nil {
		return client, errExtractClientIDFromMessageHeader(err)
	}
	return client, nil
}
func ExtractPlayerID(d rabbitmq.Delivery) (uuid.UUID, error) {
	id, ok := d.Headers[playerIDHeaderKey]
	if ok == false {
		return uuid.Nil, errExtractClientIDFromMessageHeader(errClientIDNotFound)
	}
	client, err := uuid.FromString(id.(string))
	if err != nil {
		return client, errExtractClientIDFromMessageHeader(err)
	}
	return client, nil
}
