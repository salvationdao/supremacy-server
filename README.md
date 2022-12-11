# Supremacy Gameserver

[![Staging Deployment](https://github.com/ninja-syndicate/supremacy-gameserver/actions/workflows/deploy-staging.yml/badge.svg)](https://github.com/ninja-syndicate/supremacy-gameserver/actions/workflows/deploy-staging.yml)

Supremacy gameserver for communication between the game client and various frontends

[CD Docs](.github/workflows/README.md)

## To be open sourced

- github.com/ninja-software/log_helpers -> github.com/salvationdao/log_helpers
- github.com/ninja-software/terror/v2 -> github.com/salvationdao/terror/v2
- github.com/ninja-syndicate/ws -> github.com/salvationdao/ws

### For go private modules

```shell
git config --global --add url."git@github.com:".insteadOf "https://github.com/"
export GOPRIVATE="github.com/ninja-software/*,github.com/ninja-syndicate/*"
```

### Envars

```
export GAMESERVER_DATABASE_PORT= - postgres database port
export GAMESERVER_TWITCH_EXTENSION_SECRET="" - in your twitch dev console
export GAMESERVER_GAMECLIENT_BUILD_NO="" - you must set this to the current build number for prod
export GAMESERVER_TELEGRAM_BOT_TOKEN= (in 1Password)
export GAMESERVER_STRIPE_SECRET_KEY= (in 1Password)
export GAMESERVER_STRIPE_WEBHOOK_SECRET= (in 1Password)
```

### For spinup

#### generate golang tools

```shell
make tools
```

Due to data migration, both servers must be on for a spinup process which migrates data back and forth between the two servers

```
GAMESERVER$ make db-reset
GAMESERVER$ make db-boiler
GAMESERVER$ make serve

PASSPORT-SERVER$ make db-reset
PASSPORT-SERVER$ make db-boiler
PASSPORT-SERVER$ make serve

GAMESERVER$ make sync
PASSPORT-SERVER$ make sync
```

After both servers are running (and database setup), suck data in this order:

- passport-server -> gameserver
- gameserver -> passport-server

```
cd $GAMESERVER
cd server
go run cmd/gameserver/main.go sync
```

```
cd $PASSPORTSERVER
go run cmd/platform/main.go sync
```

#### db

```shell
make docker-start docker-setup db-reset
```

#### db-boiler

For existing db, migrate up is required, to allow sqlboiler to generate boilers codes

```bash
make tools
make db-migrate
make db-boiler
```

#### manually spinup server

```shell
go run cmd/gameserver/main.go serve
```

#### load live server

```shell
make serve
```

#### run caddy

```shell
make lb
```

### Game Image Assets (Post Production)

In case you have not migrated to `000004_game_ability_changes.up.sql`, you will need to run the following command:

```
cd server
go run cmd/gameserver/main.go db --assets
```

This will seed in all the known game ability images.

### Gameserver connection details

The game client will require these headers to connect.

```shell
Sec-WebSocket-Protocol: gameserver-v1
```

### Fiat Stripe Setup

To handle payments for Stripe:

First install `stripe-cli` (see https://stripe.com/docs/stripe-cli#install)

And then run the following:

```
make serve-stripe-webhook
```

#### From Codi

````Go. Oauth2 client. Basically a proxy server connecting all the other servers (game and passport)

No UI, no user stories.

Server to server comms:

Push to Arena:

    Do game action

Receive from Arena:

    Victory Result

Push to Passport:

    Winnings
    Take money for executing game action

Receive from Twitch UI:

    Enlist requests
    Game actions```

#### From Codi 2
```
Go. Oauth2 client. Basically a proxy server connecting all the other servers (game and passport)

No UI, no user stories.

Server to server comms:

Push to Arena:
- Do game action

Receive from Arena:
- Victory Result

Push to Passport:
- Winnings
- Take money for executing game action

Receive from Twitch UI:
- Enlist requests
- Game actions

# Game Server

FEATURE:
Supremacy arena game actions should take time to execute, and allow players from one faction to cancel the game action of another faction by spending SUPs

FLOW:
- Faction #1 spends SUPs to start game action
- Takes 30 seconds to cast action
- Faction #2 spends SUPs to cancel out previous game action

IMPACT:
- Players spend SUPs
- Players create splinter groups to co-ordinate times, attacks and defenses
- Emergent player community creation

AFFECTS:
- Arena
- Game server

## Receives

- request to start match
    - Response:
        - game client build number
        - match number
        - participant war machines
            - 3 from each faction
- match started
    - data structure
        - match number
        - timestamp
    - Trigger match started event
        - Broadcast to hub clients
- war machine stat change
    - data structure
        - damaged by (war machine or action)
        - damage amount
        - weapon used
    - Trigger stat change event
        - Broadcast to hub clients
- war machine destroyed
    - data structure
        - destroyed by
        - weapon used
        - timestamp
- match ended
    - data structure
        - winner
        - remaining health
- action request (with SUP payment)
    - responds:
        - warmup time
    - trigger event to begin warmup
- action cancellation request (with SUP payment)
    - trigger event to show cancellation
- action performed
    - data structure
        - action id
        - damaged war machines
        - healed war machines
    - trigger event begin cooldown


## Send
- action
    - action id
    - action type
    - location (x,y)
    - for faction
```

````
