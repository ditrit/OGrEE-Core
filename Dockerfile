#
# Dockerfile for the prototype
#
#FROM ubuntu:latest
#LABEL author="Ziad Khalaf"
FROM alpine:latest AS builder
USER root
RUN apk add --no-cache git make musl-dev go

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin
WORKDIR /home

ADD . /home/
COPY ./resources/test/ /home/
COPY ./.env /home/
# RUN cd p3 && go mod init p3


# Install Dependencies
RUN go get -u github.com/gorilla/mux
RUN go get -u go.mongodb.org/mongo-driver
RUN go get -u github.com/dgrijalva/jwt-go
RUN go get -u github.com/joho/godotenv
RUN go get -u golang.org/x/crypto/bcrypt
RUN go get go.mongodb.org/mongo-driver/x/mongo/driver@v1.8.4
RUN apk add --no-cache python3 py3-pip
RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/community py3-pip

RUN cd /home && go build main.go 


FROM alpine:latest
USER root

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR /home

ADD . /home/
COPY ./resources/test/ /home/
COPY ./.env /home/
COPY --from=builder /home/main /home