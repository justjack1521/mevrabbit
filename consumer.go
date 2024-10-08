package mevrabbit

import (
	"context"
	"github.com/justjack1521/mevrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
	"log/slog"
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

func (s *StandardConsumer) run() {
	errCh := make(chan error, 1)
	go func() {
		err := s.actual.Run(s.standardConsumption())
		errCh <- err
		close(errCh)
	}()
	go func() {
		if err := <-errCh; err != nil {
			panic(err)
		}
	}()
}

func NewStandardConsumer(conn *rabbitmq.Conn, queue Queue, key RoutingKey, exchange Exchange, handler ConsumerHandler, opts ...func(*rabbitmq.ConsumerOptions)) (*StandardConsumer, error) {

	var consumer = &StandardConsumer{
		Queue:      queue,
		RoutingKey: key,
		Exchange:   exchange,
		handler:    handler,
	}

	var options = []func(*rabbitmq.ConsumerOptions){
		rabbitmq.WithConsumerOptionsRoutingKey(string(key)),
		rabbitmq.WithConsumerOptionsExchangeName(string(exchange)),
		rabbitmq.WithConsumerOptionsExchangeDurable,
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsExchangeKind(string(Direct)),
	}

	var fin = append(options, opts...)

	actual, err := rabbitmq.NewConsumer(conn, string(queue), fin...)
	if err != nil {
		return nil, err
	}
	consumer.actual = actual
	consumer.run()

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

func (s *StandardConsumer) WithTracing(tracer TransactionTracer) *StandardConsumer {
	if tracer != nil {
		s.handler = consumerTracingMiddleware(tracer, s.handler)
	}
	return s
}

func (s *StandardConsumer) WithNewRelic(relic *newrelic.Application) *StandardConsumer {
	if relic != nil {
		s.handler = consumerNewRelicMiddleWare(relic, s.handler)
	}
	return s
}

func (s *StandardConsumer) WithSlogging(slogger *slog.Logger) *StandardConsumer {
	if slogger != nil {
		s.handler = consumeSloggerMiddleware(slogger, s.handler)
	}
	return s
}

func (s *StandardConsumer) WithLogging(logger *logrus.Logger) *StandardConsumer {
	s.handler = consumeLoggerMiddleWare(logger, s.handler)
	return s
}

func ExtractUserID(d rabbitmq.Delivery) (uuid.UUID, error) {
	id, ok := d.Headers[userIDHeaderKey]
	if ok == false {
		return uuid.Nil, errExtractUserIDFromMessageHeader(errUserIDNotFound)
	}
	client, err := uuid.FromString(id.(string))
	if err != nil {
		return client, errExtractUserIDFromMessageHeader(err)
	}
	return client, nil
}
func ExtractPlayerID(d rabbitmq.Delivery) (uuid.UUID, error) {
	id, ok := d.Headers[playerIDHeaderKey]
	if ok == false {
		return uuid.Nil, errExtractPlayerIDFromMessageHeader(errPlayerIDNotFound)
	}
	client, err := uuid.FromString(id.(string))
	if err != nil {
		return client, errExtractPlayerIDFromMessageHeader(err)
	}
	return client, nil
}
