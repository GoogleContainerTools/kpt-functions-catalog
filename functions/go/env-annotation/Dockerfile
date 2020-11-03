FROM golang:1.14
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
RUN go build -v -o /usr/local/bin/config-function ./

FROM alpine:latest
COPY --from=0 /usr/local/bin/config-function /usr/local/bin/config-function
CMD ["config-function"]
