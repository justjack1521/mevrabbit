package mevrabbit

import (
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
	"log/slog"
)

func publisherSloggerMiddleware(slogger *slog.Logger, handler PublishHandler) PublishHandler {
	return func(ctx *PublisherContext) error {

		var entry = slogger.With(
			slog.Group("message_attr",
				slog.String("exchange", string(ctx.exchange)),
				slog.String("key", string(ctx.key)),
				slog.Int("length", len(ctx.delivery)),
			),
		)

		if err := handler(ctx); err != nil {
			entry.With("error", err.Error()).ErrorContext(ctx.Context, "message failed")
			return err
		}

		entry.InfoContext(ctx.Context, "message published")
		return nil

	}
}

func publisherLoggerMiddleWare(logger *logrus.Logger, handler PublishHandler) PublishHandler {
	return func(ctx *PublisherContext) error {

		var entry = logger.WithFields(logrus.Fields{
			"exchange":    ctx.exchange,
			"routing.key": ctx.key,
			"length":      len(ctx.delivery),
		})

		if err := handler(ctx); err != nil {
			entry.WithError(err).Error("message failed")
			return err
		}

		entry.Info("message published")
		return nil

	}
}

func publisherTracerMiddleware(tracer TransactionTracer, handler PublishHandler) PublishHandler {
	return func(ctx *PublisherContext) error {
		var segment = tracer.NewSegment(ctx.Context)
		if segment == nil {
			return handler(ctx)
		}
		defer segment.End()
		return handler(ctx)
	}
}

func publisherNewRelicMiddleware(handler PublishHandler) PublishHandler {
	return func(ctx *PublisherContext) error {
		var txn = newrelic.FromContext(ctx)
		if txn == nil {
			return handler(ctx)
		}
		segment := newrelic.MessageProducerSegment{
			StartTime:            txn.StartSegmentNow(),
			Library:              "RabbitMQ",
			DestinationType:      newrelic.MessageExchange,
			DestinationName:      string(ctx.exchange),
			DestinationTemporary: false,
		}
		defer segment.End()
		return handler(ctx)
	}
}

func consumeSloggerMiddleware(slogger *slog.Logger, handler ConsumerHandler) ConsumerHandler {
	return func(ctx *ConsumerContext) (rabbitmq.Action, error) {

		var entry = slogger.With(
			slog.Group("message_attr",
				slog.String("message_id", ctx.Delivery.MessageId),
				slog.String("exchange", ctx.Delivery.Exchange),
				slog.String("key", ctx.Delivery.RoutingKey),
			),
		)

		result, err := handler(ctx)

		if err != nil {
			entry.With("error", err.Error()).ErrorContext(ctx.Context, "failed to consume message")
		} else {
			entry.InfoContext(ctx.Context, "message consumed")
		}

		return result, err

	}
}

func consumeLoggerMiddleWare(logger *logrus.Logger, handler ConsumerHandler) ConsumerHandler {
	return func(ctx *ConsumerContext) (rabbitmq.Action, error) {
		logger.WithFields(logrus.Fields{
			"exchange":    ctx.Delivery.Exchange,
			"routing_key": ctx.Delivery.RoutingKey,
		}).Info("Message Received for Consumption")

		action, err := handler(ctx)

		defer func() {
			if err != nil {
				logger.WithFields(logrus.Fields{
					"exchange":    ctx.Delivery.Exchange,
					"routing_key": ctx.Delivery.RoutingKey,
					"action":      action,
				}).WithError(err).Error("Failed to Consume Message")
			} else {
				logger.WithFields(logrus.Fields{
					"exchange":    ctx.Delivery.Exchange,
					"routing_key": ctx.Delivery.RoutingKey,
					"action":      action,
				}).Info("Message Consumed")
			}
		}()
		return action, err
	}
}

func consumerNewRelicMiddleWare(relic *newrelic.Application, handler ConsumerHandler) ConsumerHandler {
	return func(ctx *ConsumerContext) (rabbitmq.Action, error) {

		var transaction = relic.StartTransaction("message." + ctx.Delivery.RoutingKey + ":" + ctx.Delivery.Exchange)
		transaction.AddAttribute("user.id", ctx.userID.String())
		transaction.AddAttribute("player.id", ctx.playerID.String())
		transaction.AddAttribute("message.routingKey", ctx.Delivery.RoutingKey)
		transaction.AddAttribute("message.exchange", ctx.Delivery.Exchange)
		transaction.AddAttribute("message.type", ctx.Delivery.Type)

		action, err := handler(ctx)

		if err != nil {
			transaction.NoticeError(err)
		}
		transaction.End()

		return action, err

	}
}
