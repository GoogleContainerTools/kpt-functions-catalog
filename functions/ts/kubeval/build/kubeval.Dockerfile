ARG BUILDER_IMAGE
ARG BASE_IMAGE

FROM --platform=$BUILDPLATFORM $BUILDER_IMAGE AS builder

RUN mkdir -p /home/node/app && \
    chown -R node:node /home/node/app

USER node

WORKDIR /home/node/app

# Install dependencies and cache them.
COPY --chown=node:node package*.json ./
RUN npm ci --ignore-scripts

# Build the source.
COPY --chown=node:node tsconfig.json .
COPY --chown=node:node src src
RUN npm run build && \
    npm prune --production && \
    rm -r src tsconfig.json

#############################################

FROM --platform=$BUILDPLATFORM golang:1.17-alpine3.15 AS kubeval-builder

ARG TARGETOS TARGETARCH
ARG KUBEVAL_VERSION="v0.16.1"
RUN apk update && apk add curl git
RUN git clone https://github.com/instrumenta/kubeval.git && cd kubeval && git checkout v0.16.1 && GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /usr/local/bin/kubeval ./

#############################################

FROM $BASE_IMAGE

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app
COPY --from=kubeval-builder /usr/local/bin/kubeval /usr/local/bin/kubeval
ADD jsonschema /jsonschema

ENTRYPOINT ["node", "/home/node/app/dist/kubeval_run.js"]
