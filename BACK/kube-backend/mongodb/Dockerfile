FROM mongo:latest

COPY ./init.sh /docker-entrypoint-initdb.d/init.sh
COPY ./createdb.js /home/createdb.js

RUN chmod +x /docker-entrypoint-initdb.d/init.sh


