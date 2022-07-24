# Create builder image using latest lightweight version of golang
FROM golang:alpine as builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# Install tzdata
# tzdata is required for providing the correct timezone to the production container
RUN apk update --no-cache && apk add --no-cache tzdata

# Set up the enviornment
ENV GOOS linux
ENV CGO_ENABLED 0
ENV PROJECT=docker-sql-api

# Create a user so that the image doesn't run as root
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "100001" \
    "appuser"

RUN mkdir -p /src/${PROJECT} /bin/${PROJECT}

# Set a working directory
WORKDIR /src/${PROJECT}

# Copying go.mod and go.sum files
COPY go.mod go.sum  ./

# Caching dependencies
RUN go mod download

# Copy entire source directory
COPY . .

# Build the static binary
RUN go build -o /bin/${PROJECT}/server main.go

# Create lightweight image to run the server binary
FROM alpine as production

# Installing ca-certificates
# ca-certificates is required to avoid issues with TLS
RUN apk --no-cache add ca-certificates
#RUN apk --no-cache add bind-tools

RUN mkdir -p /bin/${PROJECT} /etc/${PROJECT}

WORKDIR /root/

ENV PROJECT=docker-sql-api
ENV TZ Australia/Brisbane

# Copy the binary file built in previous stage, and .env file from source
COPY --from=builder /bin/${PROJECT}/server /bin/${PROJECT}/server
COPY --from=builder /src/${PROJECT}/.env /etc/${PROJECT}/.env

#CMD dig mysql-db
CMD /bin/${PROJECT}/server /etc/${PROJECT}/.env