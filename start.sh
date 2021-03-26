#
# Bash script to start the cockroach server for 
# testing with the prototype

killall cockroach
rm -rf ogreedb
sleep 10

cockroach start-single-node \
--insecure \
--store=ogreedb \
--listen-addr=netbox:26257 \
--http-addr=netbox:8080 \
--background

cockroach sql 	\
		--insecure \
		 --host=netbox:26257 <<EOF 
CREATE USER maxroach;
CREATE DATABASE ogreedb;
SET DATABASE = ogreedb;
GRANT ALL ON DATABASE ogreedb TO maxroach;
$(<resources/database_table_create.sql)
$(<resources/patch1.sql)
$(<resources/add_constraints.sql)
EOF