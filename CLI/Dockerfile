
FROM golang:1.19.6-bullseye AS builder
USER root

#Setup app files
WORKDIR /home
ADD CLI /home/

RUN make

FROM busybox:1.35.0-uclibc as busybox

#Final output image
FROM gcr.io/distroless/base-debian11

COPY --from=busybox /bin/sh /bin/sh

WORKDIR /home
ADD . /home/
COPY --from=builder /home/cli /home/
ENTRYPOINT ["/home/cli"]