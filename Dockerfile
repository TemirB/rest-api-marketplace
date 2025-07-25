 FROM golang:1.23.5
 WORKDIR /app

 COPY go.mod go.sum ./
 RUN go mod download

 COPY . .

 RUN go build -o server ./cmd/api

 EXPOSE 8080
 CMD ["./server"]
