FROM golang:1.9-alpine3.7

ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV GOOS linux

COPY . /go/src/transcode/transcode-delete

WORKDIR /go/src/transcode/transcode-delete

RUN apk --no-cache add curl git && \
    echo "Pulling watchdog binary from Github." && \
    curl -sSL https://github.com/openfaas/faas/releases/download/0.7.9/fwatchdog > /usr/bin/fwatchdog && \
    chmod +x /usr/bin/fwatchdog && \
    apk del curl --no-cache && \
    echo "Build binary" && \
    go get && \
    go build -o /usr/local/bin/transcode-delete

ENV fprocess="transcode-delete"
ENV write_debug="false"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD [ "fwatchdog" ]
