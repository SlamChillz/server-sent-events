package main

import (
	"fmt"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"log"
)

const streamName = "sslogs"

var streamEnvironment *stream.Environment

func init() {
	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost("localhost").
			SetPort(5552).
			SetUser("guest").
			SetPassword("guest"))
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
	fmt.Printf("env created\n")
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
	log.Printf("Name: %v", name)
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
					closeStreamConsumer <- true
					delete(clients, name)
					close(msgChan)
				default:
					msgChan <- fmt.Sprintf("%s", message.Data[0])
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
