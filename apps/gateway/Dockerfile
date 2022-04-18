FROM node:latest
WORKDIR /home/node
USER node
# RUN npm config set prefix "/tmp/npm/node_modules"
# RUN npm config set cache "/tmp/npm/cache"
# RUN npm config set tmp "/tmp/npm/tmp"
COPY --chown=node package.json ./package.json
RUN npm i
COPY --chown=node . ./
ENTRYPOINT ["npm","start"]
