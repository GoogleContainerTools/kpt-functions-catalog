FROM node:lts-alpine as builder

RUN apk add bash curl git
RUN apk update

RUN curl -fsSL -o /istio-1.6.1-linux-amd64.tar.gz https://github.com/istio/istio/releases/download/1.6.1/istio-1.6.1-linux-amd64.tar.gz && \
    tar -zxvf /istio-1.6.1-linux-amd64.tar.gz && \
    mv /istio-1.6.1/bin/istioctl /usr/local/bin/istioctl && \
    rm -f /istio-1.6.1-linux-amd64.tar.gz && \
    rm -rf /istio-1.6.1

RUN curl -fsSL -o /usr/local/bin/kpt https://storage.googleapis.com/kpt-dev/latest/linux_amd64/kpt && \
    chmod +x /usr/local/bin/kpt

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

FROM node:lts-alpine

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app

COPY --from=builder /usr/local/bin /usr/local/bin

ENV PATH /usr/local/bin:$PATH

ENTRYPOINT ["node", "/home/node/app/dist/istioctl_analyze_run.js"]
