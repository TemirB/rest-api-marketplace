 FROM golang:1.23.11-alpine3.21
 WORKDIR /app

 COPY go.mod go.sum ./
 RUN go mod download

 COPY . .

 RUN go build -o server ./cmd/api

 EXPOSE 8080
 CMD ["./server"]
