FROM golang:1.10-alpine as builder

WORKDIR /go/src/github.com/jobteaser/rediscp
COPY . /go/src/github.com/jobteaser/rediscp/

RUN apk --no-cache add git
RUN go get
RUN CGO_ENABLED=1 GOOS=linux go build -o rediscp

FROM alpine:3.7

RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/github.com/jobteaser/rediscp/rediscp /rediscp
