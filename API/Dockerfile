#
# Dockerfile for the API
#
FROM golang:latest AS builder
USER root
WORKDIR /home

RUN apt install git

COPY API/go.mod API/go.sum /home/
RUN go mod download

COPY API/ /home/
COPY .git /home/

#Build
RUN --mount=type=cache,target=/root/.cache/go-build \
    cd /home && CGO_ENABLED=0 make

FROM busybox:latest
USER root
WORKDIR /home

COPY API/resources/test/ /home/resources/test/
COPY --from=builder /home/main /home
ENTRYPOINT [ "/home/main" ]