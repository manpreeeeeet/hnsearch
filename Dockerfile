FROM node:22 AS frontend-builder

WORKDIR /app/frontend

COPY client/package*.json ./
RUN npm install

COPY client/ ./
RUN npm run build

FROM golang:1.23 AS backend-builder

WORKDIR /app/backend

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=backend-builder /app/backend/app .

COPY --from=frontend-builder /app/frontend/dist ./frontend

EXPOSE 8081

CMD ["./app", "prod"]