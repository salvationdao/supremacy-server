# supremacy-gameserver

Supremacy gameserver for communication between the game client and various frontends

### For go private modules

```shell
git config --global --add url."git@github.com:".insteadOf "https://github.com/"
export GOPRIVATE="github.com/ninja-software/*"
```

### Envars

```
export GAMESERVER_TWITCH_EXTENSION_SECRET="" - in your twitch dev console
```

### For spinup

#### generate golang tools

```shell
make tools
```

#### db

```shell
make docker-start docker-setup db-reset
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

### Gameserver connection details

The game client will require these headers to connect.

```shell
Sec-WebSocket-Protocol: gameserver-v1
```

### Certificate files for caddy

Create a folder `supremacy-gameserver/certs` and create 5 certificate files (NOTE: These certificates are only used in dev environment)

```shell
dev.supremacygame.io-ca.crt
dev.supremacygame.io-root.crt
dev.supremacygame.io.crt
server.csr
server.key
```

The content of the files can be found in Bitwarden

```shell
supremacy - dev - cert - 1
supremacy - dev - cert - 2
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
        - match number
        - participant war machines
            - 2 from each faction
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
