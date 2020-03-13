FROM alpine/helm:3.1.1

RUN apk add bash curl git
RUN apk update

RUN curl -fsSL -o /usr/local/bin/kpt https://storage.googleapis.com/kpt-dev/latest/linux_amd64/kpt
RUN chmod +x /usr/local/bin/kpt
ENV PATH /usr/local/bin:$PATH

COPY helm-template /
RUN chmod +x /helm-template

ENTRYPOINT [ "/helm-template" ]
