FROM golang:1.20 AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GIT_TERMINAL_PROMPT=1 \
    GO111MODULE=on

COPY src ${GOPATH}/mr-autocloser/src
COPY go.mod ${GOPATH}/mr-autocloser/
COPY go.sum ${GOPATH}/mr-autocloser/
COPY config ${GOPATH}/mr-autocloser/
WORKDIR ${GOPATH}/mr-autocloser
RUN go mod tidy
RUN go build -ldflags="-s -w" -o mr-autocloser ./src/

FROM scratch
COPY --from=builder go/mr-autocloser/mr-autocloser /go/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /go

ENTRYPOINT ["./mr-autocloser"]
CMD ["-config", "config.yaml"]