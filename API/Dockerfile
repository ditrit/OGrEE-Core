#
# Dockerfile for the API
#
FROM golang:latest AS builder
USER root
WORKDIR /home

COPY API/ .git* /home/

# Install Dependencies
RUN apt install git && go install .

#Build
RUN cd /home && CGO_ENABLED=0 make

FROM busybox:latest
USER root
WORKDIR /home

COPY API/resources/test/ /home/resources/test/
COPY --from=builder /home/main /home
ENTRYPOINT [ "/home/main" ]