package main

import (
	"comico/db"
	"comico/server"
	"comico/model"
)

func main() {
	db.Start()
	defer db.Stop()
	model.Init()
	server.Start()
}
