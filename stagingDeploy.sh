#!/bin/bash
# build binary
cd server/ && env GOOS=linux GOARCH=amd64 go build -o ./cmd/gameserver/gameserver ./cmd/gameserver/ && cd ../
# upload binary
scp ./server/cmd/gameserver/gameserver "root@sale.supremacy.fi:/root/gameserver"
# upload envars if needed too
#scp configs/gameserver-staging-export.env "root@sale.supremacy.fi:/home/gameserver/gameserver-staging-export.env"
#scp configs/gameserver-staging.env "root@sale.supremacy.fi:/home/gameserver/gameserver-staging.env"
# upload nginx config
scp ./config/gameserver.nginx.conf "root@sale.supremacy.fi:/home/gameserver/gameserver.conf"
# upload migrations
scp -r ./server/db/migrations "root@sale.supremacy.fi:/home/gameserver/migrations"
# upload migrate binary
scp ./bin/migrate "root@sale.supremacy.fi:/home/gameserver/migrate"
# run migrations
ssh root@sale.supremacy.fi 'cd /home/gameserver/ && source ./gameserver-staging-export.env && ./migrate -database "postgres://$GAMESERVER_DATABASE_USER:$GAMESERVER_DATABASE_PASS@$GAMESERVER_DATABASE_HOST:$GAMESERVER_DATABASE_PORT/$GAMESERVER_DATABASE_NAME?sslmode=disable" -path ./migrations drop -f'
# run migrations
ssh root@sale.supremacy.fi 'cd /home/gameserver/ && source ./gameserver-staging-export.env && ./migrate -database "postgres://$GAMESERVER_DATABASE_USER:$GAMESERVER_DATABASE_PASS@$GAMESERVER_DATABASE_HOST:$GAMESERVER_DATABASE_PORT/$GAMESERVER_DATABASE_NAME?sslmode=disable" -path ./migrations up'
# run seed
ssh root@sale.supremacy.fi 'cd /home/gameserver/ && source ./gameserver-staging-export.env && ./gameserver db'
# move binary and restart services
ssh root@sale.supremacy.fi 'chown gameserver:gameserver /root/gameserver;systemctl stop gameserver;mv /root/gameserver /home/gameserver/gameserver;systemctl start gameserver && sudo systemctl restart nginx'
