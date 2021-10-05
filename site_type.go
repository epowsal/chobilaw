package main

import (
	"github.com/epowsal/orderfile32"
)

type SiteInfo struct {
	url              string
	name             string
	catalogscript    string
	itemscript       string
	refreshscript    string
	nodowncatadbdone bool

	no_down_catalog_url_db *orderfile32.OrderFile
	no_down_item_url_db    *orderfile32.OrderFile
	item_db                *orderfile32.OrderFile
}
