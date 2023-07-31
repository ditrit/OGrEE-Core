chmod 400 /data/keyfile
chown 999:999 /data/keyfile
exec docker-entrypoint.sh mongod --bind_ip_all --replSet dbrs --keyFile /data/keyfile &
sleep 5
mongosh -u admin -p adminpassword <<EOF
var config = {
    "_id": "dbrs",
    "version": 1,
    "members": [
        {
            "_id": 1,
            "host": "dev_db:27017",
            "priority": 3
        },
    ]
};
rs.initiate(config, { force: true });
rs.status();
EOF
sleep infinity