# Install OS and dependencies to build frontend
FROM ubuntu:20.04 AS build
ARG API_URL
ARG ALLOW_SET_BACK
ARG BACK_URLS
ENV GIN_MODE=release
ENV TZ=Europe/Paris \
    DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update 
RUN apt-get install -y curl git wget unzip libgconf-2-4 gdb libstdc++6 libglu1-mesa fonts-droid-fallback lib32stdc++6 python3
RUN apt-get clean

# Download Flutter SDK from Github repo
RUN git clone --depth 1 --branch 3.7.6 https://github.com/flutter/flutter.git /usr/local/flutter

# Set flutter environment path
ENV PATH="/usr/local/flutter/bin:/usr/local/flutter/bin/cache/dart-sdk/bin:${PATH}"

# Run flutter doctor
RUN flutter doctor

# Copy files to container and build
COPY ogree_app/ /app/
WORKDIR /app/
RUN flutter pub get
RUN flutter build web --dart-define=API_URL=$API_URL --dart-define=ALLOW_SET_BACK=$ALLOW_SET_BACK --dart-define=BACK_URLS=$BACK_URLS

# Runtime image
FROM nginx:1.21.1-alpine
COPY --from=build /app/build/web /usr/share/nginx/html