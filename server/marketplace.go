package server

type MarketplaceSaleType string

const (
	MarketplaceSaleTypeBuyout       MarketplaceSaleType = "BUYOUT"
	MarketplaceSaleTypeAuction      MarketplaceSaleType = "ACTION"
	MarketplaceSaleTypeDutchAuction MarketplaceSaleType = "DUTCH_AUCTION"
)

type MarketplaceItemType string

const (
	MarketplaceItemTypeWarMachine MarketplaceItemType = "WAR_MACHINE"
)
