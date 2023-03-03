
FROM golang:1.19.6-bullseye AS builder
USER root

#Setup app files
WORKDIR /home
ADD . /home/

RUN make

#Final output image
FROM gcr.io/distroless/base-debian11
WORKDIR /home
ADD . /home/
COPY --from=builder /home/cli /home/
ENTRYPOINT ["/home/cli"]