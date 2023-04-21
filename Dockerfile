FROM golang:1.20 AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GIT_TERMINAL_PROMPT=1 \
    GO111MODULE=on

COPY src ${GOPATH}/mr-autocleaner/src
COPY go.mod ${GOPATH}/mr-autocleaner/
COPY go.sum ${GOPATH}/mr-autocleaner/
COPY config ${GOPATH}/mr-autocleaner/
WORKDIR ${GOPATH}/mr-autocleaner
RUN go mod tidy
RUN go build -ldflags="-s -w" -o mr-autocleaner ./src/

FROM scratch
COPY --from=builder go/mr-autocleaner /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /
ENTRYPOINT ["/mr-autocleaner"]