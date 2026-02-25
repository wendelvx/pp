FROM golang:1.25-alpine AS builder

# Desativa a rede para o download e usa a pasta vendor
ENV GO111MODULE=on

WORKDIR /app

# Copiamos tudo (incluindo a nova pasta vendor)
COPY . .

# O flag -mod=vendor é a chave: ele ignora a internet
RUN go build -mod=vendor -o engine main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/engine .
EXPOSE 8080
CMD ["./engine"]