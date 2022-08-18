#!/bin/bash
set -e

PACKAGE="gameserver"
read -p "Are you sure you want to rollback binary versions? (y/n)" -n 1 -r yn
case "$yn" in
    [yY] )  echo ""
            echo "Proceeding to rollback binary"
            ;;
    [nN] )  echo ""
            echo "Exiting.."
            exit
            ;;
    * )     echo "Invalid response...exiting"
            exit
            ;;
esac


systemctl stop nginx
read -t 3 -p "Stopping Nginx server"
systemctl stop gameserver
read -t 1 -p "Stopping gameserver server"
read -p "What version would you like to rollback to? (example: v3.16.10)" -r VERSION

if [ ! -d "/usr/share/ninja_syndicate/gameserver_${VERSION}" ]
then
    echo "Directory /usr/share/ninja_syndicate/gameserver_${VERSION} DOES NOT exists."
    exit 1
fi


CURVERSION=$(readlink -f ./gameserver-online)

echo "Rolling back binary version to $VERSION"

ln -Tfsv /usr/share/ninja_syndicate/gameserver_$VERSION /usr/share/ninja_syndicate/gameserver-online

date=$(date +'%Y-%m-%d-%H%M%S')
mv $CURVERSION ${CURVERSION}_BAD_${date}

LatestMigration = $(grep LatestMigration /usr/share/ninja_syndicate/gameserver-online/BuildInfo.txt | sed 's/LatestMigration=//g')

echo "Running down migrations"
source /usr/share/ninja_syndicate/gameserver-online/init/gameserver.env
sudo -u postgres ./gameserver-online/migrate -database "postgres:///$GAMESERVER_DATABASE_NAME?host=/var/run/postgresql/" -path ./migrations goto $LatestMigration

# Left commented out for now because both side will probably need to be rolledback
# systemctl start gameserver
# nginx -t && systemctl start nginx

echo "Gameserver rollbacked to Version $VERSION"