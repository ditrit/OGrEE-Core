
FROM alpine:latest AS builder
USER root
RUN apk add --no-cache git make musl-dev go

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin

WORKDIR /home
ADD .. /home/

# Install Dependencies
RUN go install .
RUN cd /home && make

FROM alpine:latest
USER root

WORKDIR /home
COPY --from=builder /home/main /home