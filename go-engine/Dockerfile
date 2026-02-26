FROM golang:1.25-alpine AS builder

# Mantém o modo de módulos ativo e garante o uso do vendor
ENV GO111MODULE=on

WORKDIR /app

# Copia todo o projeto, incluindo a nova estrutura de pastas e o vendor
COPY . .

# AJUSTE: O caminho do main.go agora é ./cmd/engine/main.go
# O binário final continuará se chamando 'engine'
RUN go build -mod=vendor -o engine ./cmd/engine/main.go

FROM alpine:latest
# Instala certificados CA caso precise bater em APIs HTTPS no futuro
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia o binário gerado no estágio anterior
COPY --from=builder /app/engine .

# Expõe a porta das métricas/healthcheck
EXPOSE 8080

CMD ["./engine"]