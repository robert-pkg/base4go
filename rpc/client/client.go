package client

import (
	"context"
)

type Client interface {
	Init() error

	Invoke(ctx context.Context, method string, args, reply any) error

	Close() error

	// Server implementation
	String() string
}
