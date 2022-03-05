PACKAGE=gameserver

# Names and Versions
DOCKER_CONTAINER=$(PACKAGE)-db

# Paths
BIN = $(CURDIR)/bin
SERVER = $(CURDIR)/server

# DB Settings
LOCAL_DEV_DB_USER=$(PACKAGE)
LOCAL_DEV_DB_PASS=dev
LOCAL_DEV_DB_HOST=localhost
LOCAL_DEV_DB_PORT=5437
LOCAL_DEV_DB_DATABASE=$(PACKAGE)
DB_CONNECTION_STRING="postgres://$(LOCAL_DEV_DB_USER):$(LOCAL_DEV_DB_PASS)@$(LOCAL_DEV_DB_HOST):$(LOCAL_DEV_DB_PORT)/$(LOCAL_DEV_DB_DATABASE)?sslmode=disable"

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
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.43.0 go get -u golang.org/x/tools/cmd/goimports
	go get -u golang.org/x/tools/cmd/goimports
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

.PHONY: db-migrate
db-migrate:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/db/migrations up

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
	cd $(SERVER) && go run cmd/gameserver/*.go db

.PHONY: db-update-assets
db-update-assets:
	cd $(SERVER) && go run cmd/gameserver/main.go db --assets

.PHONY: db-reset
db-reset: db-drop db-migrate-up-to-seed db-seed db-migrate

# make sure `make tools` is done
.PHONY: db-boiler
db-boiler:
	$(BIN)/sqlboiler $(BIN)/sqlboiler-psql --config $(SERVER)/sqlboiler.toml

.PHONY: go-mod-download
go-mod-download:
	cd $(SERVER) && go mod download

.PHONY: init
init: db-setup deps tools go-mod-tidy db-migrate

.PHONY: init-docker
init-docker: docker-start tools go-mod-tidy docker-setup db-migrate

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
	wt --window 0 --tabColor #4747E2 --title "Supremacy - Game Server" -p "PowerShell" -d ./ powershell -NoExit make serve-arelo ; split-pane --tabColor #4747E2 --title "Supremacy - Load Balancer" -p "PowerShell" -d ../supremacy-stream-site powershell -NoExit make lb ; split-pane -H -s 0.8 --tabColor #4747E2 --title "Passport Server" --suppressApplicationTitle -p "PowerShell" -d ../passport-server powershell -NoExit make serve-arelo ; split-pane --tabColor #4747E2 --title "Passport Web" -p "PowerShell" -d ../passport-web powershell -NoExit make watch ; split-pane -H -s 0.5 --tabColor #4747E2 --title "Stream Web" --suppressApplicationTitle -p "PowerShell" -d ../supremacy-stream-site powershell -NoExit npm start

.PHONY: pull
pull:
	git branch && git pull
	cd ../supremacy-stream-site && git branch && git checkout -- package-lock.json && git pull
	cd ../passport-server && git branch && git pull
	cd ../passport-web && git branch && git checkout -- package-lock.json && git pull

.PHONY: serve-test
serve-test:
	cd server && go test ./...
