FROM node:14.15-alpine3.12 as builder

ARG KUBEVAL_VERSION="0.15.0"
RUN apk add curl && \
    curl -sSLf https://github.com/instrumenta/kubeval/releases/download/${KUBEVAL_VERSION}/kubeval-linux-amd64.tar.gz | \
    tar xzf - -C /usr/local/bin

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

FROM node:14.15-alpine3.12

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app

COPY --from=builder /usr/local/bin /usr/local/bin

ENV PATH /usr/local/bin:$PATH

ENTRYPOINT ["node", "/home/node/app/dist/kubeval_run.js"]
