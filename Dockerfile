FROM golang:1.10-alpine as builder

WORKDIR /go/src/github.com/jobteaser/redis-copy
COPY . /go/src/github.com/jobteaser/redis-copy/

RUN apk --no-cache add git ca-certificates
RUN go get
RUN CGO_ENABLED=1 GOOS=linux go build -o redis-copy

FROM alpine:3.7

COPY --from=builder /go/src/github.com/jobteaser/redis-copy/redis-copy /redis-copy

