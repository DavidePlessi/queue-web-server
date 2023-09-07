# Use the official Go image as the base image
FROM golang:latest

WORKDIR /app


COPY . .
CMD ["sh", "-c", "go run main.go -t $AUTH_TOKEN -i $EXP_EL_CHK"]