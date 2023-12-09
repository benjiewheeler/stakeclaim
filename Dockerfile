# Build stage
FROM golang:1.21-alpine AS build

# change working directory
WORKDIR /opt/app

# copy go mod and sum files
ADD go.mod go.sum ./

# download all dependencies.
RUN go mod download

# copy source code file
ADD *.go ./

# compile the app
RUN CGO_ENABLED=0 GOOS=linux go build -o /opt/app/stakeclaim

#------------------------------------------------------------

# Run stage
FROM alpine:3.19

# change working directory
WORKDIR /app

# copy compiled binary
COPY --from=build /opt/app/stakeclaim ./

# mount config directory
VOLUME /config

# start
ENTRYPOINT ["/app/stakeclaim", "--config-file", "/config/config.txt"]

