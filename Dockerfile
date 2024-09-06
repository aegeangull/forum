FROM golang:latest
COPY . /app
WORKDIR /app
RUN go mod download
RUN apt-get update && apt-get install -y sqlite3
RUN go build -o forum .
EXPOSE 8080
COPY /forum.db /app/forum.db
CMD ["./forum"]
