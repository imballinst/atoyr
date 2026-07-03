#!/bin/sh
set -e

/app/server/server &

exec nginx -g "daemon off;"
