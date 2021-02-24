#
# Bash script to start the cockroach server for 
# testing with the prototype

killall cockroach
rm -rf ogreedb
sleep 10

cockroach start-single-node \
--insecure \
--store=ogreedb \
--listen-addr=localhost:26257 \
--http-addr=localhost:8080 \
--background

cockroach sql 	\
		--insecure \
		 --host=localhost:26257 <<EOF 
$(<resources/database_table_create.sql)
$(<resources/add_constraints.sql)
CREATE USER maxroach;
CREATE DATABASE ogreedb;
SET DATABASE = ogreedb;
GRANT ALL ON DATABASE ogreedb TO maxroach;
EOF