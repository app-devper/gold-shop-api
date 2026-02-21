package main

import (
	"github.com/devper-gold/gold-shop-api/app"
)

func main() {
	server := app.App{}
	server.StartApp()
}
