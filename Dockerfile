FROM golang:alpine AS builder
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build
COPY go.mod go.sum main.go ./
COPY common ./common
COPY config  ./config
COPY docs ./docs
COPY handler ./handler
COPY middleware ./middleware
COPY router ./router

RUN go mod download
RUN go build -o main .
WORKDIR /dist
RUN cp /build/main .

FROM alpine
COPY --from=builder /dist/main .
COPY localize ./localize
COPY app.env .

RUN addgroup -S 1000 && adduser -S 1000 -G 1000
RUN mkdir -p /home/1000
RUN chown -R 1000:1000 /home/1000

ENTRYPOINT ["/main"]