# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM node:10-alpine as builder

RUN mkdir -p /home/node/app && \
    chown -R node:node /home/node/app

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

# TODO(b/141115380): Remove this hack when NPM packages are published.
COPY --chown=node:node /deps/kpt-functions/kpt-functions-0.0.1.tgz /deps/kpt-functions/kpt-functions-0.0.1.tgz

# Install dependencies and cache them.
COPY --chown=node:node package.json ./
# TODO(b/141115380): Include package-lock.json from host and run 'npm ci' instead.
RUN npm install

# Build the source.
COPY --chown=node:node tsconfig.json .
COPY --chown=node:node src src
RUN npm run build && \
    npm prune --production

#############################################

FROM node:10-alpine

USER node
WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app

ENTRYPOINT ["node", "/home/node/app/dist/recommend_psp_run.js"]
