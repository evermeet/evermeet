all: docker

install:
	pnpm install --frozen-lockfile

build:
	pnpm run -r build

docker:
	sudo docker build --platform linux/amd64 -t treecz/evermeet . --progress=plain

run-docker:
	sudo docker run -p 3030:3000 treecz/evermeet

dev:
	NODE_ENV=development pm2 start ecosystem.config.js