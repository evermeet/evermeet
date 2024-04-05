
FROM node:18-alpine

#RUN mkdir -p /home/node/app/api/node_modules && mkdir -p /home/node/app/web/node_modules && chown -R node:node /home/node/app
RUN npm install pm2 -g

#USER node
WORKDIR /home/node/app
COPY config.yaml package.json ./

# INSTALL API

WORKDIR /home/node/app/packages/api
COPY packages/api/package*.json ./
RUN npm install
COPY packages/api/ ./
RUN ls -lah

# INSTALL WEB

WORKDIR /home/node/app/packages/web
COPY packages/web/package*.json ./
RUN npm install
COPY packages/web/ ./
RUN ls -lah
RUN npm run build

WORKDIR /home/node/app
COPY ecosystem.config.js ./

EXPOSE 3000
EXPOSE 3001

#CMD [ "node", "app.js" ]
CMD [ "pm2-runtime", "ecosystem.config.js" ]
