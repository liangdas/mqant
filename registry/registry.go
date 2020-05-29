// Package registry is an interface for service discovery
package registry

import (
	"errors"
)

// Registry The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type Registry interface {
	Init(...Option) error
	Options() Options
	Register(*Service, ...RegisterOption) error
	Deregister(*Service) error
	GetService(string) ([]*Service, error)
	ListServices() ([]*Service, error)
	Watch(...WatchOption) (Watcher, error)
	String() string
}

// Option Option
type Option func(*Options)

// RegisterOption RegisterOption
type RegisterOption func(*RegisterOptions)

// WatchOption WatchOption
type WatchOption func(*WatchOptions)

var (
	// DefaultRegistry 默认注册中心
	DefaultRegistry = newConsulRegistry()
	// ErrNotFound ErrNotFound
	ErrNotFound = errors.New("not found")
)

// NewRegistry 新建注册中心
func NewRegistry(opts ...Option) Registry {
	return newConsulRegistry(opts...)
}

// Register a service node. Additionally supply options such as TTL.
func Register(s *Service, opts ...RegisterOption) error {
	return DefaultRegistry.Register(s, opts...)
}

// Deregister a service node
func Deregister(s *Service) error {
	return DefaultRegistry.Deregister(s)
}

// GetService Retrieve a service. A slice is returned since we separate Name/Version.
func GetService(name string) ([]*Service, error) {
	return DefaultRegistry.GetService(name)
}

// ListServices List the services. Only returns service names
func ListServices() ([]*Service, error) {
	return DefaultRegistry.ListServices()
}

// Watch returns a watcher which allows you to track updates to the registry.
func Watch(opts ...WatchOption) (Watcher, error) {
	return DefaultRegistry.Watch(opts...)
}

// String String
func String() string {
	return DefaultRegistry.String()
}
