package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/message"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"os"
	"time"
)

func CheckErr(err error) {
	if err != nil {
		fmt.Printf("%s ", err)
		os.Exit(1)
	}
}

func handlePublishConfirm(confirms stream.ChannelPublishConfirm) {
	go func() {
		for confirmed := range confirms {
			for _, msg := range confirmed {
				if msg.IsConfirmed() {
					fmt.Printf("message %s stored \n  ", msg.GetMessage().GetData())
				} else {
					fmt.Printf("message %s failed \n  ", msg.GetMessage().GetData())
				}

			}
		}
	}()
}

func publish() {
	reader := bufio.NewReader(os.Stdin)

	//You need RabbitMQ 3.13.0 or later to run this example
	fmt.Println("Filtering example.")
	fmt.Println("Connecting to RabbitMQ streaming ...")

	// Connect to the broker ( or brokers )
	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost("localhost").
			SetPort(5552).
			SetUser("guest").
			SetPassword("guest"))
	CheckErr(err)

	err = env.DeleteStream(streamName)
	if err != nil && errors.Is(err, stream.StreamDoesNotExist) {
		// we can ignore the error if the stream does not exist
		// it will be created later
		fmt.Println("Stream does not exist. ")
	} else {
		CheckErr(err)
	}

	err = env.DeclareStream(streamName,
		&stream.StreamOptions{
			MaxLengthBytes: stream.ByteCapacity{}.GB(2),
		},
	)

	CheckErr(err)

	producer, err := env.NewProducer(streamName,
		stream.NewProducerOptions().SetFilter(
			// Here we enable the filter
			// for each message we set the filter key.
			// the filter result is a string
			stream.NewProducerFilter(func(message message.StreamMessage) string {
				return fmt.Sprintf("%s", message.GetApplicationProperties()["name"])
			})))
	CheckErr(err)

	chPublishConfirm := producer.NotifyPublishConfirmation()
	handlePublishConfirm(chPublishConfirm)

	// Send messages with the state property == New York
	send(producer, "mendy")
	time.Sleep(2 * time.Second)
	send(producer, "slam")
	time.Sleep(2 * time.Second)
	err = producer.Close()
	CheckErr(err)
	fmt.Println("Press any key to stop ")
	_, _ = reader.ReadString('\n')
	err = env.DeleteStream(streamName)
	CheckErr(err)
	err = env.Close()
	CheckErr(err)
}

func send(producer *stream.Producer, state string) {
	for i := 0; i < 100; i++ {
		msg := amqp.NewMessage([]byte(fmt.Sprintf("message %d, state %s", i, state)))
		msg.ApplicationProperties = map[string]interface{}{"name": state}
		err := producer.Send(msg)
		CheckErr(err)
	}
}
