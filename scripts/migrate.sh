#!/usr/bin/env bash

# used for dev, test and demo env

# cmd:
#    drop
#    up   [n]
#    down [n]
#    version

echo "untested"
echo "Proceed? (y/n)"
read var

if [[ $var != "y" ]]; then exit 1; fi

CLIENT="ninja_syndicate"
PACKAGE="gameserver"

if [ "$1" == "" ]; then
    echo "need cmd!" >&2
    echo "cmd:" >&2
    echo "   drop" >&2
    echo "   up   [n]" >&2
    echo "   down [n]" >&2
    echo "   version" >&2
    exit 1
fi

valid="false"
vars="up version drop"
for var in $vars; do
    if [ "$var" == "$1" ]; then
        valid="true"
    fi
done

if [ "$valid" != "true" ]; then
    echo "invalid cmd!"
    exit 2
fi

source /usr/share/${CLIENT}/${PACKAGE}_online/init/${PACKAGE}.env
source /home/ubuntu/.profile # load PGPASSWORD

# Cant use the project default user due to adjusted permisions on some tables
bin/migrate -database "postgres://postgres:${PGPASSWORD}@${PASSPORT_DATABASE_HOST}:${PASSPORT_DATABASE_PORT}/${PASSPORT_DATABASE_NAME}" -path migrations $1 $2
