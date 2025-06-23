#!/bin/bash

set -e

# Adjust to allow remove access and enable wal_level
echo "listen_addresses='*'" >>"$PGDATA/postgresql.conf"
cat >>"$PGDATA/postgresql.conf" <<-EOF
	wal_level = replica
	max_wal_senders = 10
	wal_keep_size = '1GB'
	hot_standby = on
EOF

# Allow replica user
echo "host replication ${REPLICATION_USER} 0.0.0.0/0 md5" >>"$PGDATA/pg_hba.conf"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
	CREATE ROLE ${REPLICATION_USER} WITH REPLICATION LOGIN PASSWORD '${REPLICATION_PASSWORD}';
EOSQL
