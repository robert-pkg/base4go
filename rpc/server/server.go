package server

type Server interface {
	Init() error

	// Stop the server
	Start(opts ...StartOption) error
	Stop() error

	// Server implementation
	String() string
}
