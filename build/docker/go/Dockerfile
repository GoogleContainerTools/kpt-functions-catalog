ARG BUILDER_IMAGE
ARG BASE_IMAGE


FROM $BUILDER_IMAGE
ENV CGO_ENABLED=0
WORKDIR /go/src/

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/function ./

#############################################

FROM $BASE_IMAGE
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
ENTRYPOINT ["function"]

