FROM golang:1.21-alpine AS builder

# Installer les dépendances de compilation
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod ./

# Copier le code source (nécessaire pour go mod tidy)
COPY . .

# Télécharger les dépendances et générer go.sum
RUN go mod download
RUN go mod tidy

# Compiler l'application avec les flags pour sqlite3 sur Alpine
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN CGO_ENABLED=1 GOOS=linux go build -o honeypot .

# Image finale légère
FROM alpine:latest

# Installer les dépendances runtime
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copier le binaire compilé
COPY --from=builder /app/honeypot .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Créer les dossiers nécessaires
RUN mkdir -p /root/logs /root/data /root/uploads

# Exposer les ports
EXPOSE 2222 80 443 8080

# Commande de démarrage
CMD ["./honeypot", "-config", "config.yaml"]

