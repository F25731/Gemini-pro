FROM node:22-alpine AS web-builder
WORKDIR /src/web
COPY web/package.json ./
RUN npm install
COPY web ./
RUN npm run build

FROM golang:1.22-alpine AS go-builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . ./
COPY --from=web-builder /src/web/dist ./internal/app/web/dist
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/banana-wrapper ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN adduser -D -H appuser
COPY --from=go-builder /out/banana-wrapper /app/banana-wrapper
USER appuser
EXPOSE 3000
ENTRYPOINT ["/app/banana-wrapper"]
