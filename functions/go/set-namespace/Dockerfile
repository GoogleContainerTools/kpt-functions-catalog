FROM golang:1.15
ENV CGO_ENABLED=0
WORKDIR /go/src/

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go .
COPY NamespaceTransformer.go .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /usr/local/bin/function ./

FROM alpine:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
USER nobody
CMD ["function"]
