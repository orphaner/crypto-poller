FROM golang:alpine AS build
WORKDIR /go/src/github.com/orphaner/crypto-poller/
RUN apk update && apk add git
RUN go get -d -v github.com/influxdata/influxdb/client/v2
RUN go get -d -v github.com/namsral/flag
COPY main.go httpclient.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /go/src/github.com/orphaner/crypto-poller/main .
CMD ["./main"]
