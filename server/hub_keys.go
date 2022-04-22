package server

import "github.com/ninja-syndicate/hub"

// player_abilities
const HubKeySaleAbilityDetailed = hub.HubCommandKey("SALE:ABILITY:DETAILED")
const HubKeyPlayerAbilitySubscribe = hub.HubCommandKey("PLAYER:ABILITY:SUBSCRIBE")
const HubKeySaleAbilityPriceSubscribe = hub.HubCommandKey("SALE:ABILITY:PRICE:SUBSCRIBE")
const HubKeyPlayerAbilitiesList = hub.HubCommandKey("PLAYER:ABILITIES:LIST")
const HubKeySaleAbilitiesList = hub.HubCommandKey("SALE:ABILITIES:LIST")
const HubKeyPlayerAbilitiesListUpdated = hub.HubCommandKey("PLAYER:ABILITIES:LIST:UPDATED")
const HubKeySaleAbilitiesListUpdated = hub.HubCommandKey("SALE:ABILITIES:LIST:UPDATED")
const HubKeySaleAbilityPurchase = hub.HubCommandKey("SALE:ABILITY:PURCHASE")