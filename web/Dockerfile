FROM node
WORKDIR  /app

COPY package.json .

RUN npm i -p

ARG APP
RUN mkdir ${APP}
COPY ./src/apps/${APP}/build ./${APP}

COPY ./src/apps/${APP}/package.json ./${APP}

WORKDIR /app/${APP}

ENTRYPOINT [ "npm", "run", "start" ]

