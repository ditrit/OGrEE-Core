#
# Dockerfile for the CLI
#
#LABEL author="Ziad Khalaf"
FROM python:alpine3.15 AS builder
USER root
RUN apk add --no-cache git make musl-dev go bash

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin
WORKDIR /home

ADD . /home/
#COPY /home/ziad/buildCLI/credentials/.env /home/.resources/
#COPY ./.resources/ /home/.resources/
# COPY ./.env /home/
# RUN cd p3 && go mod init p3


# Install Dependencies
RUN go install modernc.org/goyacc@latest
RUN go install github.com/blynn/nex@latest
RUN go get -u github.com/chzyer/test
RUN go get -u golang.org/x/sys

RUN make


#Final output image
FROM alpine:latest
WORKDIR /home
ADD . /home/
COPY --from=builder /home/main /home/
CMD /home/main