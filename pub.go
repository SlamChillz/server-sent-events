package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/message"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"log"
	"os"
	"os/exec"
	"strconv"
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
					log.Printf("message %s stored \n  ", msg.GetMessage().GetData())
				} else {
					log.Printf("message %s failed \n  ", msg.GetMessage().GetData())
				}

			}
		}
	}()
}

func publish(name string, streamCreated chan<- bool) {
	// Connect to the broker ( or brokers )
	RabbitPort, err := strconv.Atoi(os.Getenv("R_PORT"))
	CheckErr(err)
	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost(os.Getenv("R_HOST")).
			SetPort(RabbitPort).
			SetUser(os.Getenv("R_USER")).
			SetPassword(os.Getenv("R_PASSWORD")))
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
	streamCreated <- true
	fmt.Println("Stream created")

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
	send(producer, name)
	time.Sleep(2 * time.Second)
	err = producer.Close()
	CheckErr(err)
	//err = env.DeleteStream(streamName)
	//CheckErr(err)
	//err = env.Close()
	//CheckErr(err)
}

func send(producer *stream.Producer, name string) {
	// Start npm install process
	cmd := exec.Command("sh", "-c", "cd test; npm install; cd ..")
	//cmd := exec.Command("sh", "-c", "echo stdout; echo 1>&2 stderr")
	//cmd := exec.Command("/usr/bin/ls", "-l", ".")
	out := bytes.NewBuffer(nil)
	cmd.Stderr = out
	cmd.Stdout = out
	//stdout, err := cmd.StdoutPipe()
	//if err != nil {
	//	log.Fatalf("Failed to get stdout pipe: %v", err)
	//}
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run npm install: %v", err)
	}
	done := make(chan bool)
	defer close(done)
	go func(done chan<- bool) {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			msg := amqp.NewMessage([]byte(scanner.Text()))
			msg.ApplicationProperties = map[string]interface{}{"name": name}
			err := producer.Send(msg)
			CheckErr(err)
		}
		msg := amqp.NewMessage([]byte("done"))
		msg.ApplicationProperties = map[string]interface{}{"name": name}
		err := producer.Send(msg)
		CheckErr(err)
		done <- true
	}(done)
	//err := cmd.Wait()
	//if err != nil {
	//	fmt.Printf("npm install failed: %v\n", err)
	//}
	<-done
	log.Println("npm install completed successfully.")
}
