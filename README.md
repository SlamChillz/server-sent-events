# Server Sent Event Server
Streams realtime data to the client. Build logs are the only data currently streamed in realtime by this server.

## Prerequisites
- Linux OS
- Go 1.22 or higher

## Installation
- Clone repository: git clone https://github.com/ignitedotdev/sse.git
- Install dependencies: `go mod tidy`

## Build and Run
The sse server can be built and run both locally and in docker
- ### Local
  - Build: `go build -o sse`
  - Run: `./sse`
- ### Docker
  - Build: `docker build -t sse:local .`
  - Run: `docker run --rm sse:local`
