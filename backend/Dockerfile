FROM golang:1.27-alpine

WORKDIR /app

RUN apk add --no-cache git && go install github.com/cosmtrek/air@v1.61.2

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air"]
