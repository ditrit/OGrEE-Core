FROM mongo:6.0.9
ENV DB_NAME=AutoTest
ENV CUSTOMER_API_PASSWORD=123
ENV ADMIN_DB=admin
ENV SUPER_USER=super
ENV SUPER_PASS=superpassword
ENV MONGO_INITDB_ROOT_USERNAME=admin
ENV MONGO_INITDB_ROOT_PASSWORD=adminpassword
ENV GUARD_USER=guard
ENV GUARD_PASS=adminpassword
RUN mkdir /scripts

COPY init.sh /docker-entrypoint-initdb.d/init.sh
COPY createdb.js /home/createdb.js
COPY entrypoint.sh /scripts/entrypoint.sh

RUN chmod +x /docker-entrypoint-initdb.d/init.sh
RUN chmod +x /scripts/entrypoint.sh

ENTRYPOINT ["/scripts/entrypoint.sh"]