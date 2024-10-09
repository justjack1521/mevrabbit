package mevrabbit

import "context"

type TransactionTracer interface {
	FromContext(ctx context.Context) Transaction
	NewSegment(ctx context.Context) Segment
}

type Transaction interface {
	NoticeError(err error)
}

type Segment interface {
	AddAttribute(key string, val any)
	End()
}
