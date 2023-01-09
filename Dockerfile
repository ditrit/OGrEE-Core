#
# Dockerfile for the CLI
#
#LABEL author="Ziad Khalaf"
FROM golang:latest AS builder
USER root

#Setup build environment
RUN apt-get update && apt-get -y install python3 python3-setuptools python3-pip


#Setup app files
WORKDIR /home
ADD . /home/


#Setup build dependencies
RUN go install modernc.org/goyacc@latest
RUN go install github.com/blynn/nex@latest
RUN go get -u github.com/chzyer/test
RUN go get -u golang.org/x/sys


#Generate Binary
RUN make


#Final output image
FROM alpine:latest
WORKDIR /home
ADD . /home/
COPY --from=builder /home/main /home/
CMD /home/main