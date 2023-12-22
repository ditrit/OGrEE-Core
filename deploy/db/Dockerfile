FROM mongo:6.0.9

COPY ../docker/init.sh /docker-entrypoint-initdb.d/init.sh
COPY ./createdb.js /home/createdb.js

RUN chmod +x /docker-entrypoint-initdb.d/init.sh