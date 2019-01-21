package main

import (
	"github.com/ddidwyll/comico/db"
	"github.com/ddidwyll/comico/server"
	"github.com/ddidwyll/comico/model"
)

func main() {
	db.Start()
	defer db.Stop()
	model.Init()
	server.Start()
}
