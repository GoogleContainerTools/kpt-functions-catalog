############################################
# BUILD augeas
FROM alpine AS augeas
RUN apk update && apk add bash --no-cache --virtual=.build gcc libc-dev make bison flex readline-dev libxml2-dev git automake autoconf libtool pkgconf coreutils bash
COPY /augeas /augeas
RUN set -ex ; \
    cd augeas \
 && ./autogen.sh --prefix=/opt/augeas \
 && make \
 && make install

############################################
FROM --platform=$BUILDPLATFORM golang:1.17-alpine3.15
RUN apk add --no-cache pkgconfig libgcc libxml2 readline bash pkgconfig libxml2-dev gcc libc-dev
COPY --from=augeas /opt/augeas /opt/augeas
ENV PATH=$PATH:/opt/augeas/bin
RUN set -x ; \
    cd /opt/augeas/bin \
 && for TOOL in augcheck auggrep augload augloadone augparsediff augsed ; \
    do \
      wget https://raw.githubusercontent.com/raphink/augeas-sandbox/master/"$TOOL" ; \
      chmod +x "$TOOL" ; \
    done

RUN mkdir /etc/native-config-adaptor
RUN chmod -R 777 /etc/native-config-adaptor

WORKDIR /go/src/
ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/opt/augeas/lib/pkgconfig
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/opt/augeas/lib
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o /opt/augeas/bin/function ./
RUN chown root:root /opt/augeas/bin/function
ENTRYPOINT ["function"]


