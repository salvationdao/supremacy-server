PACKAGE=gameserver

# Names and Versions
DOCKER_CONTAINER=$(PACKAGE)-db

# Paths
BIN = $(CURDIR)/bin
SERVER = $(CURDIR)/server

# DB Settings
LOCAL_DEV_DB_USER?=$(PACKAGE)
LOCAL_DEV_DB_PASS?=dev
LOCAL_DEV_DB_HOST?=localhost
LOCAL_DEV_DB_PORT?=5437
LOCAL_DEV_DB_DATABASE?=$(PACKAGE)
DB_CONNECTION_STRING="postgres://$(LOCAL_DEV_DB_USER):$(LOCAL_DEV_DB_PASS)@$(LOCAL_DEV_DB_HOST):$(LOCAL_DEV_DB_PORT)/$(LOCAL_DEV_DB_DATABASE)?sslmode=disable"
DB_STATIC_CONNECTION_STRING="postgres://$(LOCAL_DEV_DB_USER):$(LOCAL_DEV_DB_PASS)@$(LOCAL_DEV_DB_HOST):$(LOCAL_DEV_DB_PORT)/$(LOCAL_DEV_DB_DATABASE)?sslmode=disable&x-migrations-table=static_migrations"

GITVERSION=`git describe --tags --abbrev=0`
GITHASH=`git rev-parse HEAD`
GITBRANCH=`git rev-parse --abbrev-ref HEAD`
BUILDDATE=`date -u +%Y%m%d%H%M%S`
GITSTATE=`git status --porcelain | wc -l`
REPO_ROOT=`git rev-parse --show-toplevel`

# Make Commands
.PHONY: setup-git
setup-git:
	ln -s ${REPO_ROOT}/.pre-commit ${REPO_ROOT}/.git/hooks/pre-commit

.PHONY: clean
clean:
	rm -rf deploy

.PHONY: deploy-package
deploy-prep: clean tools build
	mkdir -p deploy
	cp $(BIN)/migrate deploy/.
	cp -r ./init deploy/.
	cp -r ./configs deploy/.
	cp -r $(SERVER)/db/migrations deploy/.

.PHONY: build
.ONESHELL:
build: setup-git
	cd $(SERVER)
	pwd
	go build \
		-ldflags "-X main.Version=${GITVERSION} -X main.GitHash=${GITHASH} -X main.GitBranch=${GITBRANCH} -X main.BuildDate=${BUILDDATE} -X main.UnCommittedFiles=${GITSTATE}" \
		-gcflags=-trimpath=$(shell pwd) \
		-asmflags=-trimpath=$(shell pwd) \
		-o ../deploy/gameserver \
		cmd/gameserver/main.go

.PHONY: tools
tools: go-mod-tidy
	@mkdir -p $(BIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.43.0 go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/tools/cmd/goimports@latest
	cd $(SERVER) && go generate -tags tools ./tools/...

.PHONY: go-mod-tidy
go-mod-tidy:
	cd $(SERVER) && go mod tidy

.PHONY: docker-start
docker-start:
	docker start $(DOCKER_CONTAINER) || docker run -d -p $(LOCAL_DEV_DB_PORT):5432 --name $(DOCKER_CONTAINER) -e POSTGRES_USER=$(PACKAGE) -e POSTGRES_PASSWORD=dev -e POSTGRES_DB=$(PACKAGE) postgres:13-alpine

.PHONY: docker-stop
docker-stop:
	docker stop $(DOCKER_CONTAINER)

.PHONY: docker-remove
docker-remove:
	docker rm $(DOCKER_CONTAINER)

.PHONY: docker-setup
docker-setup:
	docker exec -it $(DOCKER_CONTAINER) psql -U $(PACKAGE) -c 'CREATE EXTENSION IF NOT EXISTS pg_trgm; CREATE EXTENSION IF NOT EXISTS pgcrypto; CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'


.PHONY: db-setup
db-setup:
	psql -h $(LOCAL_DEV_DB_HOST) -p $(LOCAL_DEV_DB_PORT) -U postgres -f init.sql

.PHONY: db-version
db-version:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/db/migrations version

.PHONY: db-drop
db-drop:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/db/migrations drop -f

.PHONY: db-drop-sync
db-drop-sync:
	$(BIN)/migrate -database $(DB_STATIC_CONNECTION_STRING) -path $(SERVER)/db/static drop -f

.PHONY: db-migrate
db-migrate:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/db/migrations up

.PHONY: db-migrate-sync
db-migrate-sync:
	$(BIN)/migrate -database $(DB_STATIC_CONNECTION_STRING) -path $(SERVER)/db/static up


.PHONY: db-migrate-sync-down-one
db-migrate-sync-down-one:
	$(BIN)/migrate -database $(DB_STATIC_CONNECTION_STRING) -path $(SERVER)/db/static down 1


.PHONY: db-migrate-down
db-migrate-down:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/db/migrations down

.PHONY: db-migrate-down-one
db-migrate-down-one:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/db/migrations down 1

.PHONY: db-migrate-up-one
db-migrate-up-one:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/db/migrations up 1

.PHONY: db-migrate-up-to-seed
db-migrate-up-to-seed:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/db/migrations up 14

.PHONY: db-prepare
db-prepare: db-drop db-migrate

.PHONY: db-seed
db-seed:
	cd $(SERVER) && go run seed/*.go db

.PHONY: db-seed-windows
db-seed-windows:
	cd $(SERVER) && go run seed/main.go seed/gameMaps.go seed/seed.go db  

.PHONY: db-update-assets
db-update-assets:
	cd $(SERVER) && go run cmd/gameserver/main.go db --assets

.PHONY: db-reset
db-reset: db-drop db-drop-sync db-migrate-sync dev-sync-data db-migrate-up-to-seed db-seed db-migrate db-boiler

.PHONY: db-reset-windows
db-reset-windows: db-drop db-drop-sync db-migrate-sync dev-sync-data-windows db-migrate-up-to-seed db-seed-windows db-migrate

# make sure `make tools` is done
.PHONY: db-boiler
db-boiler:
	$(BIN)/sqlboiler $(BIN)/sqlboiler-psql --config $(SERVER)/sqlboiler.toml

# make sure `make tools` is done
.PHONY: db-boiler-windows
db-boiler-windows:
	cd $(BIN) && sqlboiler psql --config $(SERVER)/sqlboiler.toml

.PHONY: go-mod-download
go-mod-download:
	cd $(SERVER) && go mod download

.PHONY: init
init: db-setup deps tools go-mod-tidy db-reset

.PHONY: init-docker
init-docker: docker-start tools go-mod-tidy docker-setup db-reset

.PHONY: deps
deps: go-mod-download

.PHONY: serve
serve:
	cd $(SERVER) && ${BIN}/air -c .air.toml

.PHONY: serve-arelo
serve-arelo:
	cd $(SERVER) && ${BIN}/arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -i 'tools/*' -- go run cmd/gameserver/main.go serve

.PHONY: lb
lb:
	./bin/caddy run

.PHONY: wt
wt:
	wt --window 0 --tabColor "#4747E2" --title "Supremacy - Game Server" -p "PowerShell" -d . powershell -NoExit make serve-arelo ; wt --window 0 sp --tabColor "#4747E2" --title "Supremacy - Load Balancer" -p "PowerShell" -d $(CURDIR)/../passport-web powershell -NoExit make lb-windows ; wt --window 0 mf right sp -H -s 0.8 --tabColor "#4747E2" --title "Passport Server" --suppressApplicationTitle -p "PowerShell" -d $(CURDIR)/../xsyn-services powershell -NoExit make serve-arelo ; wt --window 0 mf right sp --tabColor "#4747E2" --title "Passport Web" -p "PowerShell" -d $(CURDIR)/../passport-web powershell -NoExit npm run start:windows ; wt --window 0 mf right sp -H -s 0.5 --tabColor "#4747E2" --title "Stream Web" --suppressApplicationTitle -p "PowerShell" -d $(CURDIR)/../supremacy-play-web powershell -NoExit npm run start:windows

.PHONY: pull
pull:
	git branch && git pull
	cd ../supremacy-play-web && git branch && git checkout -- package-lock.json && git pull
	cd ../xsyn-services && git branch && git pull
	cd ../passport-web && git branch && git checkout -- package-lock.json && git pull

.PHONY: serve-test
serve-test:
	cd server && go test ./...

.PHONY: sync
sync:
	cd server && go run cmd/gameserver/main.go sync
	rm -rf ./synctool/temp-sync

.PHONY: dev-sync
dev-sync:
	cd server && go run devsync/main.go sync
	rm -rf ./synctool/temp-sync

.PHONY: dev-sync-windows
dev-sync-windows:
	cd ./server && go run ./devsync/main.go sync
	Powershell rm -r -Force ./synctool/temp-sync

.PHONY: docker-db-dump
docker-db-dump:
	mkdir -p ./tmp
	docker exec -it ${DOCKER_CONTAINER} /usr/local/bin/pg_dump -U ${LOCAL_DEV_DB_USER} > tmp/${LOCAL_DEV_DB_DATABASE}_dump.sql

.PHONY: docker-db-restore
docker-db-restore:
	ifeq ("$(wildcard tmp/$(LOCAL_DEV_DB_DATABASE)_dump.sql)","")
		$(error tmp/$(LOCAL_DEV_DB_DATABASE)_dump.sql is missing restore will fail)
	endif
		docker exec -it ${DOCKER_CONTAINER} /usr/local/bin/psql -U ${LOCAL_DEV_DB_USER} -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE pid <> pg_backend_pid() AND datname = '${LOCAL_DEV_DB_DATABASE}'"
		docker exec -it ${DOCKER_CONTAINER} /usr/local/bin/psql -U ${LOCAL_DEV_DB_USER} -d postgres -c "DROP DATABASE $(LOCAL_DEV_DB_DATABASE)"
		docker exec -i  ${DOCKER_CONTAINER} /usr/local/bin/psql -U ${LOCAL_DEV_DB_USER} -d postgres < init.sql
		docker exec -i  ${DOCKER_CONTAINER} /usr/local/bin/psql -U ${LOCAL_DEV_DB_USER} -d $(LOCAL_DEV_DB_DATABASE) < tmp/${LOCAL_DEV_DB_DATABASE}_dump.sql

.PHONY: db-dump
db-dump:
	mkdir -p ./tmp
	pg_dump -U ${LOCAL_DEV_DB_USER} > tmp/${LOCAL_DEV_DB_DATABASE}_dump.sql

.PHONY: db-restore
db-restore:
	ifeq ("$(wildcard tmp/$(LOCAL_DEV_DB_DATABASE)_dump.sql)","")
		$(error tmp/$(LOCAL_DEV_DB_DATABASE)_dump.sql is missing restore will fail)
	endif
		psql -U ${LOCAL_DEV_DB_USER} -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE pid <> pg_backend_pid() AND datname = '${LOCAL_DEV_DB_DATABASE}'"
		psql -U ${LOCAL_DEV_DB_USER} -d postgres -c "DROP DATABASE $(LOCAL_DEV_DB_DATABASE)"
		psql -U ${LOCAL_DEV_DB_USER} -d postgres < init.sql
		psql -U ${LOCAL_DEV_DB_USER} -d $(LOCAL_DEV_DB_DATABASE) < tmp/${LOCAL_DEV_DB_DATABASE}_dump.sql

.PHONE: dev-give-weapon-crate
dev-give-weapon-crate:
	curl -i -H "X-Authorization: NinjaDojo_!" -k https://api.supremacygame.io/api/give_crates/weapon/${public_address}

.PHONE: dev-give-weapon-crates
dev-give-weapon-crates:
	make dev-give-weapon-crate public_address=0xb07d36f3250f4D5B081102C2f1fbA8cA21eD87B4

.PHONE: dev-give-mech-crate
dev-give-mech-crate:
	curl -i -H "X-Authorization: NinjaDojo_!" -k https://api.supremacygame.io/api/give_crates/mech/${public_address}

.PHONE: seed-avatars
seed-avatars:
	cd ./server && go run cmd/gameserver/main.go seed-avatars

.PHONE: dev-give-mech-crates
dev-give-mech-crates:
	make dev-give-mech-crate public_address=0xb07d36f3250f4D5B081102C2f1fbA8cA21eD87B4

.PHONY: sync-data
sync-data:
	mkdir -p ./server/synctool/temp-sync
	rm -rf ./server/synctool/temp-sync/supremacy-static-data
	git clone git@github.com:ninja-syndicate/supremacy-static-data.git -b develop ./server/synctool/temp-sync/supremacy-static-data
	make sync

.PHONY: dev-sync-data
dev-sync-data:
	mkdir -p ./server/synctool/temp-sync
	rm -rf ./server/synctool/temp-sync/supremacy-static-data
	git clone git@github.com:ninja-syndicate/supremacy-static-data.git -b develop ./server/synctool/temp-sync/supremacy-static-data
	make dev-sync

.PHONY: mac-sync-data
mac-sync-data:
	cd ./server/synctool && rm -rf temp-sync && mkdir temp-sync
	cd ./server/synctool/temp-sync && git clone git@github.com:ninja-syndicate/supremacy-static-data.git -b develop
	cd ../../../
	make dev-sync

.PHONY: dev-sync-data-windows
dev-sync-data-windows:
	cd ./server/synctool && mkdir temp-sync && cd temp-sync && git clone git@github.com:ninja-syndicate/supremacy-static-data.git
	cd ../../
	make dev-sync-windows
