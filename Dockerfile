FROM golang:1.26.1-bookworm
WORKDIR /Delivery
COPY . .
RUN go mod tidy
RUN go build -o /Delivery/exe ./cmd/app/main.go
CMD ["/Delivery/exe"]