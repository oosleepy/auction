FROM golang:1.26.5

WORKDIR /auction

COPY go.mod go.sum ./ 

RUN go mod download 

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . . 

RUN go build -o auction .

EXPOSE 8080

CMD ["./auction"]
