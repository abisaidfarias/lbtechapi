FROM golang:latest

LABEL maintainer="lbtechapi"

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

ENV PORT=8080

RUN go build

RUN find . -name "*.go" -type f ! -path "./docs/*" -delete

EXPOSE $PORT

# Create .env from environment variables if USE_LOCAL_ENV is set
CMD if [ "$USE_LOCAL_ENV" = "true" ]; then \
      echo "USE_LOCAL_ENV=true" > .env && \
      echo "SECRET_KEY=$SECRET_KEY" >> .env && \
      echo "MONGO_URI=$MONGO_URI" >> .env && \
      echo "MONGO_DB=$MONGO_DB" >> .env && \
      echo "EMAIL_FROM=$EMAIL_FROM" >> .env && \
      echo "EMAIL_PASSWORD=$EMAIL_PASSWORD" >> .env && \
      echo "EMAIL_PORT=$EMAIL_PORT" >> .env && \
      echo "SMTP_CLIENTE=$SMTP_CLIENTE" >> .env && \
      echo "MONTH_INTERVAL=$MONTH_INTERVAL" >> .env; \
    fi && \
    ./lbtechapi