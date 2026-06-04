FROM golang:latest

LABEL maintainer="lbtechapi"

# Headless Chromium for move-report PDFs (chromedp / PrintToPDF). Required in Docker/EB where no browser exists.
RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    ca-certificates \
    fonts-liberation \
    && rm -rf /var/lib/apt/lists/*

# chromedp reads this in services/move_report_pdf.go when set
ENV CHROME_PATH=/usr/bin/chromium

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