FROM golang:1.13-stretch
ENV CGO_ENABLED=0
WORKDIR /go/src/

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/function ./

#############################################

FROM alpine:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function

COPY service-map.yaml /etc/generate-services-function/
RUN chmod a=r /etc/generate-services-function/service-map.yaml
ENV GENERATE_SERVICES_MAP=/etc/generate-services-function/service-map.yaml

CMD ["function"]
