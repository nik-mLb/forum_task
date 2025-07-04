# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /usr/src/app

# Download dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy and build the application
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o forum-app .

# Runtime stage
FROM postgres:15-alpine

# Environment variables
ENV PORT=5000
ENV POSTGRES_DB=forum
ENV POSTGRES_USER=forum
ENV POSTGRES_PASSWORD=forum
ENV PGDATA=/var/lib/postgresql/data/pgdata

EXPOSE $PORT 5432

# Copy initialization scripts
COPY db.sql .
COPY --from=builder /usr/src/app/forum-app /usr/local/bin/

# Create a custom startup script
RUN echo "#!/bin/sh" > /start.sh && \
    echo "docker-entrypoint.sh postgres &" >> /start.sh && \
    echo "while ! pg_isready -U $POSTGRES_USER -d $POSTGRES_DB; do sleep 1; done" >> /start.sh && \
    echo "forum-app" >> /start.sh && \
    chmod +x /start.sh

CMD ["/start.sh"]