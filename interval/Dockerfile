FROM golang:latest

RUN mkdir /app
WORKDIR /app

# Install dependencies
RUN go get -u gopkg.in/mgo.v2
RUN go get -u goji.io
RUN go get -u github.com/hashicorp/consul/api
RUN go get -u github.com/streadway/amqp

# Add source files
ADD . /app/

# Build
RUN go build -race -o main


# REST API
EXPOSE 8080

ENTRYPOINT [ "/app/main" ]