FROM node:lts-alpine as builder

RUN apk add bash curl git && apk update

ARG HELM_VERSION="v3.3.0"
RUN curl -fsSL -o /helm-${HELM_VERSION}-linux-amd64.tar.gz https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    tar -zxvf /helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    mv /linux-amd64/helm /usr/local/bin/helm

RUN curl -fsSL -o /usr/local/bin/kpt https://storage.googleapis.com/kpt-dev/latest/linux_amd64/kpt && \
    chmod +x /usr/local/bin/kpt

RUN mkdir -p /home/node/app && \
    chown -R node:node /home/node/app

USER node

WORKDIR /home/node/app

# Install dependencies and cache them.
COPY --chown=node:node package*.json ./
# Make rw package for sops package.
# TODO: Please remove next line when https://github.com/GoogleContainerTools/kpt/issues/1026 is done
COPY --chown=node:node @types @types
RUN npm ci --ignore-scripts

# Build the source.
COPY --chown=node:node tsconfig.json .
COPY --chown=node:node src src
RUN npm run build && \
    npm prune --production && \
    rm -r src tsconfig.json

#############################################

FROM node:lts-alpine

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

# Use /home/node because user node has write permission in this directory.
# https://github.com/nodejs/docker-node/issues/740
WORKDIR /home/node

COPY --from=builder /home/node/app /home/node

COPY --from=builder /usr/local/bin /usr/local/bin

ENV PATH /usr/local/bin:$PATH
ENV HELM_PATH_CACHE /var/cache
ENV HELM_CONFIG_HOME /tmp/helm/config
ENV HELM_CACHE_HOME /tmp/helm/cache

ENTRYPOINT ["node", "/home/node/dist/helm_template_run.js"]
