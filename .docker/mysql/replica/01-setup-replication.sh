#!/bin/bash

until mysqladmin ping -h "mysql-master" --silent; do
	echo "Waiting for master node..."
	sleep 2
done

mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" <<EOSQL
  CHANGE REPLICATION SOURCE TO
    SOURCE_HOST='mysql-master',
    SOURCE_USER='replicator',
    SOURCE_PASSWORD='${MYSQL_ROOT_PASSWORD}',
    SOURCE_AUTO_POSITION=1,
    GET_SOURCE_PUBLIC_KEY=1;

  START REPLICA;
EOSQL

mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "SHOW REPLICA STATUS\G"
