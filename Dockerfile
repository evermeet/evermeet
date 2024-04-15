
FROM node:20-alpine

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"

RUN npm install pnpm -g
RUN pnpm install pm2 -g

WORKDIR /home/node/app
COPY package.json pnpm-*.yaml ecosystem.config.js ./

COPY packages ./packages
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm run install-ci

RUN pnpm run build

EXPOSE 3000
EXPOSE 3001

ENV NODE_ENV production
CMD [ "pm2-runtime", "ecosystem.config.js" ]