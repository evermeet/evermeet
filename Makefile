all: docker

install:
	pnpm install --frozen-lockfile

build:
	pnpm run -r build

docker:
	sudo docker build -t treecz/deluma . --progress=plain

run-docker:
	sudo docker run -p 3030:3000 treecz/deluma