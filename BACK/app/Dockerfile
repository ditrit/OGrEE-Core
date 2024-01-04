FROM golang:1.20-alpine3.16 AS build

# Set the Current Working Directory inside the container
WORKDIR /build

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .
	
RUN go mod download

COPY . .


# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -o kube-admin -ldflags "-X main.version=${VERSION} -X 'main.build=$(date)'" .

# Start fresh from a smaller image
FROM alpine:3.16

RUN apk add curl
ENV KUBE_VERSION=v1.28.0
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$KUBE_VERSION/bin/linux/amd64/kubectl; \
    chmod +x ./kubectl;\
    mv ./kubectl /usr/bin/kubectl

ENV VERSION=3.12.3

# ENV BASE_URL="https://storage.googleapis.com/kubernetes-helm"
ENV BASE_URL="https://get.helm.sh"

RUN case `uname -m` in \
        x86_64) ARCH=amd64; ;; \
        armv7l) ARCH=arm; ;; \
        aarch64) ARCH=arm64; ;; \
        ppc64le) ARCH=ppc64le; ;; \
        s390x) ARCH=s390x; ;; \
        *) echo "un-supported arch, exit ..."; exit 1; ;; \
    esac && \
    apk add --update --no-cache wget git curl bash yq && \
    wget ${BASE_URL}/helm-v${VERSION}-linux-${ARCH}.tar.gz -O - | tar -xz && \
    mv linux-${ARCH}/helm /usr/bin/helm && \
    chmod +x /usr/bin/helm && \
    rm -rf linux-${ARCH}
    
#Add netbox repo
RUN helm repo add bootc https://charts.boo.tc

WORKDIR /bin

COPY --from=build /build/kube-admin .
COPY --from=build /build/.env .
COPY --from=build /build/swagger.json .
COPY --from=build /build/handlers handlers
COPY --from=build /build/helm helm
COPY --from=build /build/netbox netbox


# Run the binary program produced by `go install`
CMD ["./kube-admin", "-kube"]