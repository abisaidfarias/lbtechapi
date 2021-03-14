FROM golang:latest
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .

EXPOSE 7080 
EXPOSE 27017 

CMD ["/app/main"]
