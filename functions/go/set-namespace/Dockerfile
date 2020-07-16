FROM golang:1.14
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
RUN go build -v -o /usr/local/bin/set-namespace ./

FROM alpine:latest
COPY --from=0 /usr/local/bin/set-namespace /usr/local/bin/set-namespace
CMD ["set-namespace"]
