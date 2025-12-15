FROM golang:1.21-alpine AS builder

# Installer les dépendances de compilation
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le code source
COPY . .

# Compiler l'application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o honeypot .

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
RUN mkdir -p /root/logs

# Exposer les ports
EXPOSE 2222 8080

# Commande de démarrage
CMD ["./honeypot", "-config", "config.yaml"]

