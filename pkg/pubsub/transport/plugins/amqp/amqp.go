package amqp

import (
	"context"
	"github.com/kopaygorodsky/brigadier/pkg/pubsub/transport"
	"github.com/kopaygorodsky/brigadier/pkg/pubsub/transport/pkg"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"log"
	"sync"
)

func NewTransport(url string, logger *log.Logger) transport.Transport {
	return &amqpTransport{
		url:    url,
		logger: logger,
	}
}

type amqpTransport struct {
	url        string
	connection *amqp.Connection
	receivingChan *amqp.Channel
	sendingChan *amqp.Channel
	logger     *log.Logger
}

func (t *amqpTransport) Connect(ctx context.Context) error {
	conn, err := amqp.Dial(t.url)
	if err != nil {
		return errors.WithStack(err)
	}

	sendingCh, err := conn.Channel()

	if err != nil {
		return errors.WithStack(err)
	}

	t.connection = conn
	t.sendingChan = sendingCh

	receivingChan, err := conn.Channel()

	if err != nil {
		return errors.WithStack(err)
	}

	t.receivingChan = receivingChan

	return nil
}

func (t *amqpTransport) CreateTopic(ctx context.Context, topic transport.Topic) error {
	if t.connection == nil {
		if err := t.Connect(ctx); err != nil {
			return errors.WithStack(err)
		}
	}

	amqpTopic, topicConv := topic.(amqpTopic)

	if !topicConv {
		return errors.Errorf("Supplied topic is not an instance of amqp.Topic")
	}

	err := t.sendingChan.ExchangeDeclare(
		amqpTopic.Name(),
		"topic",
		amqpTopic.durable,
		amqpTopic.autoDelete,
		amqpTopic.internal,
		amqpTopic.noWait,
		nil,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (t *amqpTransport) CreateQueue(ctx context.Context, q transport.Queue, qbs ...transport.QueueBind) error {
	if t.connection == nil {
		if err := t.Connect(ctx); err != nil {
			return errors.WithStack(err)
		}
	}

	queue, queueConv := q.(amqpQueue)

	if !queueConv {
		return errors.Errorf("Supplied Queue is not an instance of amqp.amqpQueue")
	}

	var queueBinds []amqpQueueBind

	for _, item := range qbs {
		queueBind, queueBindConv := item.(amqpQueueBind)

		if !queueBindConv {
			return errors.Errorf("One of supplied QueueBinds is not an instance of amqp.amqpQueueBind")
		}

		queueBinds = append(queueBinds, queueBind)
	}

	_, err := t.sendingChan.QueueDeclare(
		queue.Name(),
		queue.durable,
		queue.autoDelete,
		queue.exclusive,
		queue.noWait,
		nil,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, qb := range queueBinds {
		err := t.sendingChan.QueueBind(
			queue.Name(),
			qb.BindingKey(),
			qb.DestinationTopic(),
			qb.noWait,
			nil,
		)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (t *amqpTransport) Send(ctx context.Context, outboundPkg pkg.OutboundPkg, options ...transport.SendOpts) error {
	if t.connection == nil {
		if err := t.Connect(ctx); err != nil {
			return errors.WithStack(err)
		}
	}

	sendOptions := &SendOptions{}

	for _, opt := range options {
		if err := opt(sendOptions); err != nil {
			return errors.WithStack(err)
		}
	}

	err := t.sendingChan.Publish(
		outboundPkg.Destination().DestinationTopic,
		outboundPkg.Destination().RoutingKey,
		sendOptions.Mandatory,
		sendOptions.Immediate,
		amqp.Publishing{
			Headers:     outboundPkg.Headers(),
			ContentType: outboundPkg.ContentType(),
			Body:        outboundPkg.Payload(),
		},
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (t *amqpTransport) Consume(ctx context.Context, queues []transport.Queue, options ...transport.ConsumeOpts) (<-chan pkg.IncomingPkg, error) {
	if t.connection == nil {
		if err := t.Connect(ctx); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	consumeOptions := &ConsumeOptions{}

	for _, opt := range options {
		if err := opt(consumeOptions); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	income := make(chan pkg.IncomingPkg)

	consumersWait := sync.WaitGroup{}

	for _, q := range queues {
		consumersWait.Add(1)
		go func(queue transport.Queue) {
			defer consumersWait.Done()

			ch, err := t.connection.Channel()

			if err != nil {
				t.logger.Print(err)
				return
			}

			defer func() {
				if err := ch.Close(); err != nil {
					t.logger.Print(err)
				}
			}()

			msgs, err := ch.Consume(
				queue.Name(),
				consumeOptions.Consumer,
				consumeOptions.AutoAck,
				consumeOptions.Exclusive,
				consumeOptions.NoLocal,
				consumeOptions.NoWait,
				nil,
			)

			if err != nil {
				t.logger.Print(err)
				return
			}

			for {
				select {
				case msg, open := <-msgs:
					if !open {
						t.logger.Printf("Amqp consumer closed channel for queue %s", queue.Name())
						return
					}
					inPkg := pkg.NewAmqpIncomingPackage(msg, msg.MessageId, queue.Name())

					income <- inPkg
				case <-ctx.Done():
					t.logger.Printf("Canceled context. Stopped consuming queue %s", queue.Name())
					return
				}
			}
		}(q)
	}

	go func() {
		consumersWait.Wait()
		close(income)
	}()

	return income, nil
}

func (t *amqpTransport) Disconnect(ctx context.Context) error {
	if t.connection == nil || t.channel == nil {
		return nil
	}

	if err := t.channel.Close(); err != nil {
		return errors.WithStack(err)
	}

	if err := t.connection.Close(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}