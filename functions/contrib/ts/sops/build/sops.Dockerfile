FROM node:14.15-alpine3.12 as builder

RUN apk add bash curl git && apk update

ARG SOPS_VERSION="v3.7.1"
RUN curl -fsSL -o /usr/local/bin/sops https://github.com/mozilla/sops/releases/download/${SOPS_VERSION}/sops-${SOPS_VERSION}.linux && \
    chmod +x /usr/local/bin/sops

RUN curl -fsSL -o /usr/local/bin/kpt https://storage.googleapis.com/kpt-dev/latest/linux_amd64/kpt && \
    chmod +x /usr/local/bin/kpt

RUN mkdir -p /home/node/app && \
    chown -R node:node /home/node/app

USER node

WORKDIR /home/node/app

# Install dependencies and cache them.
COPY --chown=node:node package*.json ./
# Make rw package work
COPY --chown=node:node @types @types

RUN npm ci --ignore-scripts

# Copy the source.
COPY --chown=node:node tsconfig.json .
COPY --chown=node:node src src

# Build the source.
RUN npm run build && \
    npm prune --production && \
    rm -r src tsconfig.json

#############################################

FROM node:14.15-alpine3.12

RUN apk add git gnupg

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app

COPY --from=builder /usr/local/bin /usr/local/bin

ENV PATH /usr/local/bin:$PATH
ENV GNUPGHOME /tmp
ENV XDG_CONFIG_HOME /tmp

ENTRYPOINT ["node", "/home/node/app/dist/sops_run.js"]
