FROM node:lts-alpine as builder

RUN apk add bash curl git && apk update

ARG KUSTOMIZE_VERSION="v3.8.1"
RUN curl -fsSL -o /kustomize-${KUSTOMIZE_VERSION}-linux-amd64.tar.gz https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_amd64.tar.gz && \
    tar -zxvf /kustomize-${KUSTOMIZE_VERSION}-linux-amd64.tar.gz && \
    mv kustomize /usr/local/bin/kustomize

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

RUN apk add git docker-cli

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app

COPY --from=builder /usr/local/bin /usr/local/bin

ENV PATH /usr/local/bin:$PATH

ENTRYPOINT ["node", "/home/node/app/dist/kustomize_build_run.js"]
