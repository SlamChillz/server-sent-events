package main

import (
	"bytes"
	"fmt"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"log"
	"os"
	"strconv"
)

const streamName = "sslogs"

var streamEnvironment *stream.Environment

func init() {
	RabbitPort, err := strconv.Atoi(os.Getenv("R_PORT"))
	CheckErr(err)
	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost(os.Getenv("R_HOST")).
			SetPort(RabbitPort).
			SetUser(os.Getenv("R_USER")).
			SetPassword(os.Getenv("R_PASSWORD")))
	if err != nil {
		log.Fatalf("Failed to connect to message broker due to: %v", err)
	}
	err = env.DeclareStream(streamName,
		&stream.StreamOptions{
			MaxLengthBytes: stream.ByteCapacity{}.GB(2),
		})
	if err != nil {
		log.Fatalf("Failed to declare stream in message broker due to: %v", err)
	}
	streamEnvironment = env
}

func cretePostFilter(name string) func(*amqp.Message) bool {
	return func(msg *amqp.Message) bool {
		return msg.ApplicationProperties["name"] == name
	}
}

type StreamConsumer struct {
	Name string
	*stream.Consumer
	Filter         *stream.ConsumerFilter
	MessageHandler func(ctx stream.ConsumerContext, message *amqp.Message)
}

func NewStreamConsumer(name string, done <-chan bool, closeStreamConsumer chan<- bool) *StreamConsumer {
	filter := stream.NewConsumerFilter([]string{name}, true, cretePostFilter(name))
	consumer, err := streamEnvironment.NewConsumer(
		streamName,
		func(done <-chan bool, closeStreamConsumer chan<- bool) func(ctx stream.ConsumerContext, message *amqp.Message) {
			return func(ctx stream.ConsumerContext, message *amqp.Message) {
				msgChan, ok := clients[name]
				if !ok {
					return
				}
				select {
				case <-done:
					// At this point the client might still have realtime logs coming in
					// So decide what to do considering all cases
					// Should we save the logs till the client reconnects? If yes, track the offset of the consumer stream
					// Should we just discard the logs?
					closeStreamConsumer <- true
				default:
					var buffer bytes.Buffer
					for _, slice := range message.Data {
						buffer.Write(slice)
					}
					msgChan <- fmt.Sprintf("%s", buffer.String())
				}
			}
		}(done, closeStreamConsumer),
		stream.NewConsumerOptions().
			SetOffset(stream.OffsetSpecification{}.First()).
			SetFilter(filter),
	)
	if err != nil {
		log.Fatalf("Failed to create new stream consumer due to: %v", err)
	}
	return &StreamConsumer{
		Name:     name,
		Filter:   filter,
		Consumer: consumer,
	}
}
