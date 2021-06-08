FROM golang:1.15-alpine3.12
ENV CGO_ENABLED=0
WORKDIR /go/src/

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/validate ./

#############################################

FROM alpine:3.12
COPY --from=0 /usr/local/bin/validate /usr/local/bin/validate
ENTRYPOINT ["validate"]
