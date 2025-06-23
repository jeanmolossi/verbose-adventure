#!/bin/bash

set -e

if [ -z "$(ls -A "$PGDATA")" ]; then
	echo "=== Initializing réplica via pg_basebackup ==="
	export PGPASSWORD=${REPLICATION_PASSWORD}

	pg_basebackup \
		-h "${MASTER_HOST}" \
		-p "${MASTER_PORT}" \
		-D "$PGDATA" \
		-U "${REPLICATION_USER}" \
		-v -P \
		--wal-method=stream \
		-R

	# Enable standby mode
	touch "$PGDATA/standy.signal"

	cat >>"$PGDATA/postgresql.auto.conf" <<-EOF
		    primary_conninfo = 'host=${MASTER_HOST} port=${MASTER_PORT} user=${REPLICATION_USER} password=${REPLICATION_PASSWORD}'
	EOF
fi

exec docker-entrypoint.sh postgres
