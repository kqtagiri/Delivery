FROM golang:1.26.1-bookworm AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/exe ./cmd/app/main.go

FROM alpine:3.20
WORKDIR /Delivery
COPY --from=build /build/exe .
CMD ["/Delivery/exe"]