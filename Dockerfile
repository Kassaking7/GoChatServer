FROM golang:1.20
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app/main ./main.go
RUN chmod +x /app/main
ENTRYPOINT ["/app/main"]
