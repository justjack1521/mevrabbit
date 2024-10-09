package mevrabbit

import "context"

type TransactionTracer interface {
	NewTransaction(ctx context.Context, name string) (context.Context, Transaction)
	NewSegment(ctx context.Context) Segment
}

type Transaction interface {
	Segment
	NoticeError(err error)
}

type Segment interface {
	AddAttribute(key string, val any)
	End()
}
