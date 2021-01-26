FROM docker

# Install gcc, bash and make
RUN apk add --no-cache build-base
RUN apk add --no-cache bash
RUN apk add --no-cache make
RUN apk add --no-cache curl

# Install golang
COPY --from=golang:1.15-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

# Install node and npm
RUN apk add --no-cache --update nodejs npm
