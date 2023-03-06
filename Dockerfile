#
# Dockerfile for the API
#
FROM golang:latest AS builder
USER root
WORKDIR /home

ADD . /home/

# Install Dependencies
RUN go get -u \
github.com/gorilla/mux \
go.mongodb.org/mongo-driver \
github.com/dgrijalva/jwt-go \
github.com/joho/godotenv \
golang.org/x/crypto/bcrypt

RUN go get \
go.mongodb.org/mongo-driver/x/mongo/driver@latest \
go.mongodb.org/mongo-driver/x/mongo/driver/auth@latest \
go.mongodb.org/mongo-driver/x/mongo/driver/ocsp@latest \
go.mongodb.org/mongo-driver/x/mongo/driver/topology@latest

#Build
RUN cd /home && make


FROM busybox:latest
USER root
WORKDIR /home

ADD . /home/
COPY ./resources/test/ /home/
COPY ./.env /home/
COPY --from=builder /home/main /home