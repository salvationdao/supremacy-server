package server

import "github.com/ninja-syndicate/hub"

const HubKeyPlayerAbilitySubscribe = hub.HubCommandKey("PLAYER:ABILITY:SUBSCRIBE")
const HubKeySaleAbilityPriceSubscribe = hub.HubCommandKey("SALE:ABILITY:PRICE:SUBSCRIBE")
const HubKeyPlayerAbilitiesList = hub.HubCommandKey("PLAYER:ABILITIES:LIST")
const HubKeySaleAbilitiesList = hub.HubCommandKey("SALE:ABILITIES:LIST")
const HubKeySaleAbilityPurchase = hub.HubCommandKey("SALE:ABILITY:PURCHASE")
