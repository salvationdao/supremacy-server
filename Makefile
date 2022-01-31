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

BUILDDATE=`date -u +%Y%m%d%H%M%S`
GITSTATE=`git status --porcelain | wc -l`

# Make Commands
.PHONY: clean
clean:
	rm -rf deploy

.PHONY: deploy-package
deploy-prep: clean tools build
	mkdir -p deploy/migrations/
	cp $(BIN)/migrate deploy/.
	cp -r $(SERVER)/db/migrations deploy/migrations/.

.PHONY: build
build:
	cd $(SERVER) && go build \
		-ldflags "-X main.BuildDate=${BUILDDATE} -X main.UnCommittedFiles=$(GITSTATE)" \
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
	psql -U postgres -f init.sql

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

.PHONY: db-prepare
db-prepare: db-drop db-migrate

.PHONY: db-seed
db-seed:
	cd $(SERVER) && go run cmd/gameserver/main.go db

.PHONY: db-reset
db-reset: db-drop db-migrate db-seed

.PHONY: go-mod-download
go-mod-download:
	cd $(SERVER) && go mod download

.PHONY: init
init: db-setup deps tools go-mod-tidy db-migrate

.PHONY: init-docker
init-docker: docker-start deps tools go-mod-tidy docker-setup db-migrate

.PHONY: deps
deps: go-mod-download

.PHONY: serve
serve:
	cd $(SERVER) && ${BIN}/air -c .air.toml

.PHONY: serve-arelo
serve-arelo:
	cd $(SERVER) && ${BIN}/arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -i 'tools/*' -- go run cmd/platform/main.go serve

.PHONY: lb
lb:
	./bin/caddy run

.PHONY: wt
wt:
	wt --window 0 --tabColor #4747E2 --title "Boilerplate - Server" -p "PowerShell" -d ./server powershell -NoExit "${BIN}/arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -i 'tools/*' -- go run cmd/platform/main.go serve" ; split-pane --tabColor #4747E2 --title "Boilerplate - Load Balancer" -p "PowerShell" -d ./ powershell -NoExit make lb ; split-pane -H -s 0.8 --tabColor #4747E2 --title "Boilerplate - Admin Frontend" --suppressApplicationTitle -p "PowerShell" -d ./web powershell -NoExit "$$env:BROWSER='none' \; npm run admin-start" ; split-pane -H -s 0.5 --tabColor #4747E2 --title "Boilerplate - Public Frontend" --suppressApplicationTitle -p "PowerShell" -d ./web powershell -NoExit "$$env:BROWSER='none' \; npm run public-start"

.PHONY: serve-test
serve-test:
	cd server && go test ./...
