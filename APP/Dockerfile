# Install OS and dependencies to build frontend
FROM ubuntu:20.04 AS build
ENV GIN_MODE=release
ENV TZ=Europe/Paris \
    DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update 
RUN apt-get install -y curl git wget unzip libgconf-2-4 gdb libstdc++6 libglu1-mesa fonts-droid-fallback python3
RUN apt-get clean

# Download Flutter SDK from Github repo
RUN git clone --depth 1 --branch 3.24.1 https://github.com/flutter/flutter.git /usr/local/flutter

# Set flutter environment path
ENV PATH="/usr/local/flutter/bin:/usr/local/flutter/bin/cache/dart-sdk/bin:${PATH}"

# Run flutter doctor
RUN flutter doctor

# Copy files to container and build
COPY APP/ /app/
COPY API/models/schemas/ /API/models/schemas/
WORKDIR /app/
RUN flutter pub get
RUN flutter build web 

# Runtime image
FROM nginx:1.24.0-alpine
COPY --from=build /app/build/web /usr/share/nginx/html