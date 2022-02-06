#
# Dockerfile for the CLI
#
#LABEL author="Ziad Khalaf"
FROM python:alpine3.15
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
RUN go get -u github.com/chzyer/test
RUN go get -u golang.org/x/sys
RUN go get -u golang.org/x/tools/cmd/goyacc
RUN go get -u github.com/blynn/nex
RUN make


#WORKDIR $GOPATH

#CMD ["make"]

#RUN cd /home && go build main.go 
CMD /home/main