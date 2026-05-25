#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
    SELECT 'CREATE DATABASE clicksafe'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'clicksafe')\gexec
    SELECT 'CREATE DATABASE clicksafe_max_adapter'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'clicksafe_max_adapter')\gexec
EOSQL
