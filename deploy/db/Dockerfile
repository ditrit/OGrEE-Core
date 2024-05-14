FROM mongo:6.0.9

COPY docker/init.sh /docker-entrypoint-initdb.d/init.sh
COPY db/createdb.js /home/createdb.js
COPY docker/entrypoint.sh /scripts/entrypoint.sh

RUN chmod +x /docker-entrypoint-initdb.d/init.sh

ENTRYPOINT ["sh", "/scripts/entrypoint.sh", "mongo-db"]