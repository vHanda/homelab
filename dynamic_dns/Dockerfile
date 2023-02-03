FROM golang:1.18 as build

WORKDIR /app
COPY . /app

RUN go mod download
RUN CGO_ENABLED=0 go build -o dns main.go

FROM alpine:latest as server

WORKDIR /app/
COPY --from=build /app/dns /app/dns
RUN chmod +x ./dns

CMD [ "/app/dns" ]
