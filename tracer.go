package mevrabbit

import "context"

type TransactionTracer interface {
	NewRabbitMQTransaction(ctx context.Context, name string) (context.Context, Transaction)
	NewRabbitMQSegment(ctx context.Context, exchange Exchange) Segment
}

type Transaction interface {
	Segment
	NoticeError(err error)
}

type Segment interface {
	AddAttribute(key string, val any)
	End()
}
