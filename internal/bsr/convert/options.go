// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package convert

// getOpts - iterate the inbound Options and return a struct
func getOpts(opt ...Option) options {
	opts := getDefaultOptions()
	for _, o := range opt {
		o(&opts)
	}
	return opts
}

// Option - how Options are passed as arguments
type Option func(*options)

// options = how options are represented
type options struct {
	withChannelId string
}

func getDefaultOptions() options {
	return options{}
}

// WithChannelId provides and option to specify the channelId
func WithChannelId(id string) Option {
	return func(o *options) {
		o.withChannelId = id
	}
}