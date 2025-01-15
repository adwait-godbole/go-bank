#!/bin/sh

# the script will exit immediately if any command returns a non-zero status.
set -e

echo "run db migration"
source /app/app.env # Read the values as environment variables as Viper has not yet come into picture at this point
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@" # takes all parameters passed to the script and run it.