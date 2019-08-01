package servers

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

type grpcTransportMananger struct {
	addr   string
	server *grpc.Server
}

func (t *grpcTransportMananger) Start() <-chan error {
	errC := make(chan error)

	go func() {
		nl, err := net.Listen("tcp", t.addr)
		if err != nil {
			errC <- err
			return
		}

		errC <- t.server.Serve(nl)
	}()

	return (<-chan error)(errC)
}

func (t *grpcTransportMananger) Stop() error {
	t.server.GracefulStop()
	return nil
}

// WithGRPCServer will run a gRPC-server when the Server is Run
func WithGRPCServer(addr string, server *grpc.Server) Option {
	if server == nil {
		panic("must provide both a net listener and GRPC server to WithGRPCServer")
	}

	return func(o *serverOptions) error {
		o.transports = append(o.transports, &grpcTransportMananger{
			addr:   addr,
			server: server,
		})
		return nil
	}
}

type httpTransportMananger struct {
	addr   string
	server *http.Server
}

func (t *httpTransportMananger) Start() <-chan error {
	errC := make(chan error)

	go func() {
		nl, err := net.Listen("tcp", t.addr)
		if err != nil {
			errC <- err
			return
		}

		errC <- t.server.Serve(nl)
	}()

	return (<-chan error)(errC)
}

func (t *httpTransportMananger) Stop() error {
	return t.server.Shutdown(context.Background())
}

// WithHTTPServer will run a http server when the Server is Run
func WithHTTPServer(addr string, server *http.Server) Option {
	if server == nil {
		panic("must provide both a net listener and HTTP server to WithHTTPServer")
	}

	return func(o *serverOptions) error {
		o.transports = append(o.transports, &httpTransportMananger{
			addr:   addr,
			server: server,
		})
		return nil
	}
}

type (
	customTransportMananger struct {
		impl customTransportManangerImpl
	}

	customTransportManangerImpl interface {
		Run() error
		Shutdown() error
	}
)

func (t *customTransportMananger) Start() <-chan error {
	errC := make(chan error)

	go func() {
		errC <- t.impl.Run()
	}()

	return (<-chan error)(errC)
}

func (t *customTransportMananger) Stop() error {
	return t.impl.Shutdown()
}

// WithCustom will run a custom transport.
func WithCustom(addr string, impl customTransportManangerImpl) Option {
	if impl == nil {
		panic("must provide a customTransportManangerImpl to WithCustom")
	}

	return func(o *serverOptions) error {
		o.transports = append(o.transports, &customTransportMananger{
			impl: impl,
		})
		return nil
	}
}
