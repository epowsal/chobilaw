package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed icon.png
var iconct []byte
var appicon = &fyne.StaticResource{
	StaticName:    "icon.png",
	StaticContent: iconct,
}
