package dialer

import "golang.org/x/net/proxy"

type clientOptions struct {
	proxyDialer proxy.Dialer
}

// ClientOption ...
type ClientOption interface {
	apply(*clientOptions)
}

type functionClientOption struct {
	f func(*clientOptions)
}

// implement ClientOption
func (f *functionClientOption) apply(opts *clientOptions) {
	f.f(opts)
}

// WithClientProxy ...
func WithClientProxy(dialer proxy.Dialer) ClientOption {
	return &functionClientOption{
		f: func(options *clientOptions) {
			options.proxyDialer = dialer
		},
	}
}
