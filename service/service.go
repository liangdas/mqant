package service

import (
	"sync"
	"time"

	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/server"
)

// NewService NewService
func NewService(opts ...Option) Service {
	return newService(opts...)
}

type service struct {
	opts Options

	once sync.Once
}

func newService(opts ...Option) Service {
	options := newOptions(opts...)

	return &service{
		opts: options,
	}
}

func (s *service) run(exit chan bool) {
	if s.opts.RegisterInterval <= time.Duration(0) {
		return
	}

	t := time.NewTicker(s.opts.RegisterInterval)

	for {
		select {
		case <-t.C:
			err := s.opts.Server.ServiceRegister()
			if err != nil {
				log.Warning("service run Server.Register error: ", err)
			}
		case <-exit:
			t.Stop()
			return
		}
	}
}

// Init initialises options. Additionally it calls cmd.Init
// which parses command line flags. cmd.Init is only called
// on first Init.
func (s *service) Init(opts ...Option) {
	// process options
	for _, o := range opts {
		o(&s.opts)
	}

	s.once.Do(func() {
		// save user action

	})
}

func (s *service) Options() Options {
	return s.opts
}

func (s *service) Server() server.Server {
	return s.opts.Server
}

func (s *service) String() string {
	return "mqant"
}

func (s *service) Start() error {
	for _, fn := range s.opts.BeforeStart {
		if err := fn(); err != nil {
			return err
		}
	}

	if err := s.opts.Server.Start(); err != nil {
		return err
	}

	if err := s.opts.Server.ServiceRegister(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStart {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) Stop() error {
	var gerr error

	for _, fn := range s.opts.BeforeStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	if err := s.opts.Server.ServiceDeregister(); err != nil {
		return err
	}

	if err := s.opts.Server.Stop(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	return gerr
}

func (s *service) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	// start reg loop
	ex := make(chan bool)
	go s.run(ex)

	//ch := make(chan os.Signal, 1)
	//signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	// wait on kill signal
	//case <-ch:
	// wait on context cancel
	case <-s.opts.Context.Done():
	}

	// exit reg loop
	close(ex)
	return s.Stop()
}
