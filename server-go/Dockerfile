FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server-go .

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache tzdata ca-certificates
COPY --from=build /app/server-go .
EXPOSE 8081
ENTRYPOINT ["./server-go"]
