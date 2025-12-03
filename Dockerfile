FROM golang:latest

LABEL maintainer="lbtechapi"

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

# Copy environment file for DEV (Parameter Store values baked in for Docker limitation)
COPY .env.docker .env

ENV PORT=8080
ENV USE_LOCAL_ENV=true

RUN go build

RUN find . -name "*.go" -type f ! -path "./docs/*" -delete

EXPOSE $PORT

CMD ["./lbtechapi"]