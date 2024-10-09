package mevrabbit

import (
	"context"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
	"log/slog"
)

type PublisherContext struct {
	context.Context
	userID   uuid.UUID
	playerID uuid.UUID
	delivery []byte
	key      RoutingKey
	exchange Exchange
}

type PublishHandler func(ctx *PublisherContext) error

type StandardPublisher struct {
	exchange Exchange
	kind     ExchangeKind
	closed   bool
	actual   *rabbitmq.Publisher
	handler  PublishHandler
}

func (s *StandardPublisher) Close() {
	if s.closed == true {
		return
	}
	s.closed = true
	s.actual.Close()
}

func (s *StandardPublisher) Publish(ctx context.Context, bytes []byte, user, player uuid.UUID, key RoutingKey) error {
	var cont = &PublisherContext{
		Context:  ctx,
		userID:   user,
		playerID: player,
		delivery: bytes,
		key:      key,
		exchange: s.exchange,
	}
	return s.handler(cont)
}

func NewUserPublisher(conn *rabbitmq.Conn, options ...func(publisherOptions *rabbitmq.PublisherOptions)) *StandardPublisher {
	actual, err := newPublisher(conn, User, Direct, options...)
	if err != nil {
		panic(err)
	}
	return newStandardPublisher(User, Direct, actual)
}
func NewClientPublisher(conn *rabbitmq.Conn, options ...func(publisherOptions *rabbitmq.PublisherOptions)) *StandardPublisher {
	actual, err := newPublisher(conn, Client, Direct, options...)
	if err != nil {
		panic(err)
	}
	return newStandardPublisher(Client, Direct, actual)
}

func NewSocialPublisher(conn *rabbitmq.Conn, options ...func(publisherOptions *rabbitmq.PublisherOptions)) *StandardPublisher {
	actual, err := newPublisher(conn, Social, Direct, options...)
	if err != nil {
		panic(err)
	}
	return newStandardPublisher(Social, Direct, actual)
}

func NewRankingPublisher(conn *rabbitmq.Conn, options ...func(publisherOptions *rabbitmq.PublisherOptions)) *StandardPublisher {
	actual, err := newPublisher(conn, Ranking, Direct, options...)
	if err != nil {
		panic(err)
	}
	return newStandardPublisher(Ranking, Direct, actual)
}

func NewGamePublisher(conn *rabbitmq.Conn, options ...func(publisherOptions *rabbitmq.PublisherOptions)) *StandardPublisher {
	actual, err := newPublisher(conn, Game, Direct, options...)
	if err != nil {
		panic(err)
	}
	return newStandardPublisher(Game, Direct, actual)
}

func (s *StandardPublisher) publish(ctx *PublisherContext) error {
	var options = []func(*rabbitmq.PublishOptions){
		rabbitmq.WithPublishOptionsExchange(string(s.exchange)),
		rabbitmq.WithPublishOptionsHeaders(IdentityPublishingTable(ctx.userID, ctx.playerID)),
	}
	return s.actual.PublishWithContext(ctx, ctx.delivery, []string{string(ctx.key)}, options...)
}

func (s *StandardPublisher) WithSlogging(slogger *slog.Logger) *StandardPublisher {
	if slogger != nil {
		s.handler = publisherSloggerMiddleware(slogger, s.handler)
	}
	return s
}

func (s *StandardPublisher) WithLogging(logger *logrus.Logger) *StandardPublisher {
	s.handler = publisherLoggerMiddleWare(logger, s.handler)
	return s
}

func (s *StandardPublisher) WithTracing(tracer TransactionTracer) *StandardPublisher {
	if tracer != nil {
		s.handler = publisherTracerMiddleware(tracer, s.handler)
	}
	return s
}

func (s *StandardPublisher) WithNewRelic() *StandardPublisher {
	s.handler = publisherNewRelicMiddleware(s.handler)
	return s
}

func newStandardPublisher(exchange Exchange, kind ExchangeKind, actual *rabbitmq.Publisher) *StandardPublisher {
	var std = &StandardPublisher{
		exchange: exchange,
		kind:     kind,
		actual:   actual,
	}
	std.handler = std.publish
	return std
}

func newPublisher(conn *rabbitmq.Conn, exchange Exchange, kind ExchangeKind, options ...func(publisherOptions *rabbitmq.PublisherOptions)) (*rabbitmq.Publisher, error) {
	options = append(
		options,
		rabbitmq.WithPublisherOptionsExchangeName(string(exchange)),
		rabbitmq.WithPublisherOptionsExchangeKind(string(kind)),
		rabbitmq.WithPublisherOptionsExchangeDurable,
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	)
	return rabbitmq.NewPublisher(conn, options...)
}

func IdentityPublishingTable(user, player uuid.UUID) rabbitmq.Table {
	return rabbitmq.Table{
		userIDHeaderKey:   user.String(),
		playerIDHeaderKey: player.String(),
	}
}
