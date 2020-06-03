FROM node:10-alpine as builder

RUN apk add bash curl git
RUN apk update

ENV HELM_LATEST_VERSION="v3.2.1"
RUN curl -fsSL -o /helm-${HELM_LATEST_VERSION}-linux-amd64.tar.gz https://get.helm.sh/helm-${HELM_LATEST_VERSION}-linux-amd64.tar.gz && \
    tar -zxvf /helm-${HELM_LATEST_VERSION}-linux-amd64.tar.gz && \
    mv /linux-amd64/helm /usr/local/bin/helm && \
    rm -f /helm-${HELM_LATEST_VERSION}-linux-amd64.tar.gz && \
    rm -rf /linux-amd64

RUN curl -fsSL -o /usr/local/bin/kpt https://storage.googleapis.com/kpt-dev/latest/linux_amd64/kpt && \
    chmod +x /usr/local/bin/kpt

RUN mkdir -p /home/node/app && \
    chown -R node:node /home/node/app

USER node

WORKDIR /home/node/app

# Install dependencies and cache them.
COPY --chown=node:node package*.json ./
RUN npm ci

# Build the source.
COPY --chown=node:node tsconfig.json .
COPY --chown=node:node src src
RUN npm run build && \
    npm prune --production && \
    rm -r src tsconfig.json

#############################################

FROM node:10-alpine

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app

COPY --from=builder /usr/local/bin /usr/local/bin

ENV PATH /usr/local/bin:$PATH
ENV HELM_PATH_CACHE /var/cache

ENTRYPOINT ["node", "/home/node/app/dist/helm_template_run.js"]
