# ---- build web ----
FROM node:22-alpine AS web
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ---- build binary ----
FROM golang:1.25-alpine AS go
WORKDIR /app
# Cache dependencies before copying source
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /app/web/build ./web/build
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o evermeet .

# ---- final image ----
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go /app/evermeet ./evermeet
EXPOSE 7331 4001
VOLUME /data
ENTRYPOINT ["/app/evermeet"]
CMD ["--config", "/data/evermeet.toml", "--data", "/data"]
