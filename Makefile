build:
	CGO_ENABLED=0 \
    GOOS=linux \
    go build -buildvcs=false -o sse
.PHONY: build

rabbitmq:
	@docker run -it --rm --name rabbitmq -p 5552:5552 -p 15672:15672 -p 5672:5672  \
            -e RABBITMQ_SERVER_ADDITIONAL_ERL_ARGS='-rabbitmq_stream advertised_host localhost' \
            rabbitmq:3.13

rabbitmq-stream:
	@docker exec rabbitmq rabbitmq-plugins enable rabbitmq_stream rabbitmq_stream_management

run: clean build
	./sse
.PHONY: run

clean:
	@rm sse | tee /dev/stderr
.PHONY: clean
