# Build the server from source
FROM golang:1.22.5 AS build-stage

WORKDIR /server

COPY go.mod ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
          -a \
          -buildvcs=false \
          -o /server/sse

# Deploy the server binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /server

COPY --from=build-stage /server/see /server/sse

EXPOSE 8080

USER sse:sse

ENTRYPOINT ["/server/sse"]
