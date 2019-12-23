FROM node:10-alpine as builder

RUN mkdir -p /home/node/app && \
    chown -R node:node /home/node/app

USER node

WORKDIR /home/node/app

# TODO(b/141115380): Remove once packages are published to public registry.
COPY --chown=node:node .npmrc .

# Install dependencies and cache them.
COPY --chown=node:node package.json ./
RUN npm install

# Build the source.
COPY --chown=node:node tsconfig.json .
COPY --chown=node:node src src
RUN npm run build && \
    npm prune --production && \
    rm -r src tsconfig.json .npmrc

#############################################

FROM node:10-alpine

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app

ENTRYPOINT ["node", "/home/node/app/dist/hydrate_anthos_team_run.js"]
