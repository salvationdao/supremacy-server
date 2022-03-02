#!/usr/bin/env bash
#
# USAGE
#
# Set all the variables inside the script, make sure you chmod +x it
#
#     version_change target_dir
#
# If your version/tag doesn't match, the script will exit with error.

CLIENT="ninja_syndicate"
PACKAGE="gameserver"



VER=$(grep -oP 'Version=\K[0-9]+' /usr/share/${CLIENT}/${PACKAGE}_online/BuildInfo.txt || echo "0")
YMDHMS=$(date +'%Y%m%d%H%M%S')
DBDIR="/usr/share/${CLIENT}/${PACKAGE}_online/db_copy"
mkdir -p $DBDIR
DBFILE="$DBDIR/$PACKAGE_$YMDHMS.sql"

# Start the change over

source ${PACKAGE}_online/init/${PACKAGE}.env

source /home/ubuntu/.profile # load PGPASSWORD

# Cant use the project default user due to adjusted permisions on some tables
pg_dump --dbname="$PASSPORT_DATABASE_NAME" --host="$PASSPORT_DATABASE_HOST" --port="$PASSPORT_DATABASE_PORT" --username="postgres" > ${DBFILE}

if [ ! -s "${DBFILE}" ]; then
    echo "db copy is zero size" >&2
    exit 2
fi
