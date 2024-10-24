FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN apt update && apt install -y libopus-dev

RUN go mod download

COPY . .

# Copy the certificates into the image
COPY certs/mumble-key.pem certs/
COPY certs/mumble-cert.pem certs/

RUN CGO_ENABLED=1 go build -o main .

CMD ["./main"]