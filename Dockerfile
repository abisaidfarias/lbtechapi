FROM golang:latest

LABEL maintainer="lbtechapi"

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

ENV PORT=8080
ENV ENVIRONMENT=prod

RUN go build

RUN find . -name "*.go" -type f ! -path "./docs/*" -delete

EXPOSE $PORT

CMD ["./lbtechapi"]