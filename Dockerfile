# Install OS and dependencies to build frontend
FROM ubuntu:20.04
ENV GIN_MODE=release
ENV TZ=Europe/Paris \
    DEBIAN_FRONTEND=noninteractive

RUN apt-get update 
RUN apt-get install -y curl git wget unzip libgconf-2-4 gdb libstdc++6 libglu1-mesa fonts-droid-fallback lib32stdc++6 python3
RUN apt-get clean

# Download Flutter SDK from Flutter Github repo
RUN git clone --depth 1 --branch 3.3.10 https://github.com/flutter/flutter.git /usr/local/flutter

# Set flutter environment path
ENV PATH="/usr/local/flutter/bin:/usr/local/flutter/bin/cache/dart-sdk/bin:${PATH}"

# Run flutter doctor
RUN flutter doctor

# Copy files to container and build
COPY ogree_app/ /app/
WORKDIR /app/
RUN flutter pub get
RUN flutter build web

# Record the exposed port 5000 and run frontend
EXPOSE 5000
COPY server.sh /app/
COPY ogree_app/.env /app/build/web/
RUN ["chmod", "+x", "/app/server.sh"]
ENTRYPOINT [ "/app/server.sh"]
