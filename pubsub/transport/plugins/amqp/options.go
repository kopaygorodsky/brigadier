package amqp

import (
	"github.com/go-foreman/foreman/pubsub/transport"
	"github.com/pkg/errors"
)

type ConsumeOptions struct {
	Exclusive     bool
	NoLocal       bool
	NoWait        bool
	PrefetchCount int
}

func convertConsumeOptsType(options interface{}) (*ConsumeOptions, error) {
	opts, ok := options.(*ConsumeOptions)

	if !ok {
		return nil, errors.Errorf("SuppliedOption must be amqp.ConsumeOptions")
	}

	return opts, nil
}

func convertSendOptsType(options interface{}) (*SendOptions, error) {
	opts, ok := options.(*SendOptions)

	if !ok {
		return nil, errors.Errorf("SuppliedOption must be amqp.SendOptions")
	}

	return opts, nil
}

func WithQosPrefetchCount(limit int) transport.ConsumeOpts {
	return func(options interface{}) error {
		opts, err := convertConsumeOptsType(options)

		if err != nil {
			return errors.WithStack(err)
		}
		opts.PrefetchCount = limit
		return nil
	}
}

func WithExclusive() transport.ConsumeOpts {
	return func(options interface{}) error {
		opts, err := convertConsumeOptsType(options)

		if err != nil {
			return errors.WithStack(err)
		}

		opts.Exclusive = true

		return nil
	}
}

func WithNoLocal() transport.ConsumeOpts {
	return func(options interface{}) error {
		opts, err := convertConsumeOptsType(options)

		if err != nil {
			return errors.WithStack(err)
		}

		opts.NoLocal = true

		return nil
	}
}

func WithNoWait() transport.ConsumeOpts {
	return func(options interface{}) error {
		opts, err := convertConsumeOptsType(options)

		if err != nil {
			return errors.WithStack(err)
		}

		opts.NoWait = true

		return nil
	}
}

type SendOptions struct {
	Mandatory bool
	Immediate bool
}

func WithMandatory() transport.SendOpts {
	return func(options interface{}) error {
		opts, err := convertSendOptsType(options)

		if err != nil {
			return errors.WithStack(err)
		}

		opts.Mandatory = true

		return nil
	}
}

func WithImmediate() transport.SendOpts {
	return func(options interface{}) error {
		opts, err := convertSendOptsType(options)

		if err != nil {
			return errors.WithStack(err)
		}

		opts.Mandatory = true

		return nil
	}
}
