
FROM node:18-alpine

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"

RUN apk add make
RUN npm install pnpm -g
RUN pnpm install pm2 -g

WORKDIR /home/node/app
COPY package.json pnpm-*.yaml ecosystem.config.js Makefile config.yaml ./

COPY packages ./packages
RUN --mount=type=cache,id=pnpm,target=/pnpm/store make install


RUN ls -lah .
RUN make build

EXPOSE 3000
EXPOSE 3001

CMD [ "pm2-runtime", "ecosystem.config.js" ]