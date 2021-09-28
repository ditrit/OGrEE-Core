#
# Dockerfile for the prototype
#
#FROM ubuntu:latest
#LABEL author="Ziad Khalaf"
FROM alpine:latest
USER root
RUN apk add --no-cache git make musl-dev go

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin
WORKDIR /home

ADD . /home/
COPY ./.env /home/
# RUN cd p3 && go mod init p3


# Install Dependencies
RUN go get -u github.com/gorilla/mux
RUN go get -u go.mongodb.org/mongo-driver
RUN go get -u github.com/dgrijalva/jwt-go
RUN go get -u github.com/joho/godotenv
RUN go get -u golang.org/x/crypto/bcrypt

#WORKDIR $GOPATH

#CMD ["make"]

RUN cd /home && go build main.go 