
# Copy in preview-hierarchy
FROM golang:1.16-alpine3.13
ENV CGO_ENABLED=0
WORKDIR /go/src/

COPY third_party ./third_party/
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build preview-hierarchy
RUN go build -o /usr/local/bin/function ./

# Final collection of artifacts
FROM alpine:3.13
COPY template.tmpl /
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
ENTRYPOINT ["function"]
