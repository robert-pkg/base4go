package server

type Server interface {
	Init(opts ...Option) error

	// Stop the server
	Start(opts ...StartOption) error
	Stop() error

	// Server implementation
	String() string
}
