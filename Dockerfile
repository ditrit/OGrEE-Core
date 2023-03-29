#
# Dockerfile for the API
#
FROM golang:latest AS builder
USER root
WORKDIR /home

ADD . /home/

# Install Dependencies
RUN apt install git && go install .

#Build
RUN cd /home && make


FROM busybox:latest
USER root
WORKDIR /home

ADD . /home/
COPY ./resources/test/ /home/
COPY --from=builder /home/main /home
ENTRYPOINT [ "/home/main" ]