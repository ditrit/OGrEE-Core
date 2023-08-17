FROM golang:1.18.2-alpine3.16 AS build

# Set the Current Working Directory inside the container
WORKDIR /build

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .
	
RUN go mod download

COPY . .


# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -o ogree-bff  -ldflags "-X main.version=${VERSION} -X 'main.build=$(date)'" .

# Start fresh from a smaller image
FROM alpine:3.16

WORKDIR /bin

COPY --from=build /build/ogree-bff .
COPY --from=build /build/api.json .
COPY --from=build /build/swagger.json .


# Run the binary program produced by `go install`
CMD ["./ogree-bff"]