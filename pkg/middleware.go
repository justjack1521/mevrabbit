package pkg

import (
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
)

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
