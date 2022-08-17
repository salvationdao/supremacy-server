#!/bin/bash
set -e

PACKAGE="passport-api"
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

echo "Stopping nginx service"
systemctl stop nginx
sleep 3
echo "Stopping passport service"
systemctl stop passport
sleep 1
read -p "What version would you like to rollback to? (example: v3.16.10)" -r VERSION

if [ ! -d "/usr/share/ninja_syndicate/passport-api_${VERSION}" ]
then
    echo "Directory /usr/share/ninja_syndicate/passport-api_${VERSION} DOES NOT exists."
    exit 1
fi

CURVERSION=$(readlink -f ./passport-online)

echo "Rolling back binary version to $VERSION"

ln -Tfsv /usr/share/ninja_syndicate/passport-api_$VERSION /usr/share/ninja_syndicate/passport-online

date=$(date +'%Y-%m-%d-%H%M%S')
mv $CURVERSION ${CURVERSION}_BAD_${date}

LatestMigration=$(grep LatestMigration /usr/share/ninja_syndicate/passport-online/BuildInfo.txt | sed 's/LatestMigration=//g')

echo "Running down migrations"
source /usr/share/ninja_syndicate/passport-online/init/passport-staging.env
sudo -u postgres ./passport-online/migrate -database "postgres:///$PASSPORT_DATABASE_NAME?host=/var/run/postgresql/" -path ${CURVERSION}_BAD_${date}/migrations goto $LatestMigration

# Left commented out for now because both side will probably need to be rolledback
# systemctl start passport
# nginx -t && systemctl start nginx

echo "Passport rollbacked to Version $VERSION"