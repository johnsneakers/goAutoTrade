package main

import (
	"goAutoTrade/services"
	"goAutoTrade/services/bitmex"
)

func main() {
	services.RunService(bitmex.NewService())
}
