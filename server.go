package servers

import (
	"time"

	"github.com/legalweb/healthz/healthz"
)

type (
	Server struct {
		// addr to bind the internal HTTP server which has /healthz and /metrics endpoints.
		addr string
		// transports are all the servers we will be handling.
		transports []transportStateManager
		// healthz configuration
		healthz *healthz.Healthz
	}

	serverOptions struct {
		addr              string
		transports        []transportStateManager
		healthzComponents []*healthz.Component
		healthzMeta       healthz.Meta
	}

	transportStateManager interface {
		Start() <-chan error
		Stop() error
	}

	Option func(*serverOptions) error
)

func New(opts ...Option) (*Server, error) {
	config := &serverOptions{
		addr:              ":8001",
		transports:        []transportStateManager{},
		healthzComponents: []*healthz.Component{},
		healthzMeta:       healthz.Meta{},
	}

	for _, o := range opts {
		if err := o(config); err != nil {
			return nil, err
		}
	}

	return &Server{
		addr:       config.addr,
		transports: config.transports,
		healthz:    healthz.New(config.healthzMeta, config.healthzComponents...),
	}, nil
}

// Run will start the internal HTTP server for health and metrics
// along with other transports that is registered on the transports
// member.
// When a transport gives an error back we will Close all the other
// transports and send the error to the channel returned.
func (s *Server) Run() <-chan error {
	errC := make(chan error)
	total := len(s.transports)
	internalErrorOffset := total
	transportErrs := make([]<-chan error, internalErrorOffset+1)
	for i := 0; i < internalErrorOffset; i++ {
		transportErrs[i] = make(chan error)
	}
	transportErrs[internalErrorOffset] = s.runInternalHTTP()

	for offset, transport := range s.transports {
		transportErrs[offset] = transport.Start()
	}
	s.healthz.Started()

	go func() {
		var transportErr error
		died := -1

		for died == -1 {
			for offset, transportErrC := range transportErrs {
				select {
				case err := <-transportErrC:
					transportErr = err
					died = offset
					goto closeServers
				default:
				}
			}

			time.Sleep(1 * time.Second)
		}

	closeServers:
		for offset, transport := range s.transports {
			if offset == died {
				continue
			}

			transport.Stop()
		}

		errC <- transportErr
	}()

	return (<-chan error)(errC)
}
