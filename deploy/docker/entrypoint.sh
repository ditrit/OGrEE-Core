openssl rand -base64 768 > /data/keyfile
chmod 400 /data/keyfile && chown 999:999 /data/keyfile
exec docker-entrypoint.sh mongod --bind_ip_all --replSet rs0 --keyFile /data/keyfile &
sleep 10
mongosh -u admin -p adminpassword <<EOF
rs.initiate({"_id": "rs0", "members": [{"_id": 1,"host": "$1:27017"}]});
rs.status();
EOF
sleep infinity