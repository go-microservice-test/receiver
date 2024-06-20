FROM golang:1.22-alpine AS build

# working directory in the container
WORKDIR /app

# copy files that contain dependencies
COPY go.mod go.sum ./

# download dependencies
RUN go mod download

# copy everything into container
COPY . .

# run the application
RUN go build -o main .

# new stage
FROM alpine:latest

# working directory in the container
WORKDIR /app

# copy prebuilt binary file
COPY --from=build /app/main .

# expose port 3000
EXPOSE 3000

# run the executable
CMD ["./main"]
