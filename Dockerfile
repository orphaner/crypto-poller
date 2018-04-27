FROM golang:alpine AS build
WORKDIR /go/src/github.com/orphaner/crypto-poller/
COPY main.go .
RUN apk update && apk add git
RUN go get -d -v github.com/influxdata/influxdb/client/v2 \
  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /go/src/github.com/orphaner/crypto-poller/main .
CMD ["./main"]
