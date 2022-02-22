#!/usr/bin/env bash
#
# USAGE
#
# Set all the variables inside the script, make sure you chmod +x it
#
#     version_change target_dir
#
# If your version/tag doesn't match, the script will exit with error.

TARGET=$(pwd)/$1
CLIENT="ninja_syndicate"
PACKAGE="gameserver"

cd /usr/share/$CLIENT

if [ -z "$1" ]; then
  echo "" >&2
  echo " USAGE" >&2
  echo "" >&2
  echo "" >&2
  echo " to bring an existing version online:" >&2
  echo "" >&2
  echo "     version_change target_dir" >&2
  echo "" >&2
  echo " If the directory doesn't exist and a tar does then it will be untared" >&2
  exit 1
fi;

if [ ! -d $TARGET ] ; then
  if [ -f $TARGET.tar.gz ];
    then tar -xvf $TARGET.tar.gz;
    else
      echo "Nither '$TARGET' or '$TARGET.tar.gz' was found in '/usr/share/$CLIENT'" >&2
      exit 1
  fi;
fi

VER=$(grep -oP 'Version=\K[0-9]+' /usr/share/${CLIENT}/${PACKAGE}_online/BuildInfo.txt || echo "0")
YMDHMS=$(date +'%Y%m%d%H%M%S')
DBDIR="/usr/share/${CLIENT}/${PACKAGE}_online/db_copy"
mkdir -p $DBDIR
DBFILE="$DBDIR/$PACKAGE_$YMDHMS.sql"

# Start the change over

source ${PACKAGE}_online/init/${PACKAGE}.env
cp ${PACKAGE}_online/init/${PACKAGE}.env $TARGET/init/${PACKAGE}.env

source /home/ubuntu/.profile # load PGPASSWORD

# Cant use the project default user due to adjusted permisions on some tables
pg_dump --dbname="$GAMESERVER_DATABASE_NAME" --host="$GAMESERVER_DATABASE_HOST" --port="$GAMESERVER_DATABASE_PORT" --username="postgres" > ${DBFILE}

if [ ! -s "${DBFILE}" ]; then
    echo "db copy is zero size" >&2
    exit 2
fi

echo "Proceed with migrations? (y/N)"
read PROCEED
if [[ $var != "y" ]]; then exit 1; fi

systemctl stop ${PACKAGE}
$TARGET/migrate -database "postgres://postgres:${PGPASSWORD}@${GAMESERVER_DATABASE_HOST}:${GAMESERVER_DATABASE_PORT}/${GAMESERVER_DATABASE_NAME}" -path $TARGET/migrations up

ln -Tfsv $TARGET $(pwd)/${PACKAGE}_online

# Ensure ownership
chown -R ${PACKAGE}:${PACKAGE} .

systemctl daemon-reload && systemctl start ${PACKAGE}
nginx -t && nginx -s reload
