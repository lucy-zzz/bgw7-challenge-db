package main

import (
	"app/internal/application"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

func main() {
	// env
	// ...

	// app
	// - config
	cfg := &application.ConfigApplicationDefault{
		Db: &mysql.Config{
			User:   "root",
			Passwd: "SENTINELA!",
			Net:    "tcp",
			Addr:   "localhost:3306",
			DBName: "fantasy_products",
		},
		Addr: "127.0.0.1:8080",
	}
	app := application.NewApplicationDefault(cfg)
	// - set up
	err := app.SetUp()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = app.InsertCustomersJSON()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = app.InsertInvoicesJSON()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = app.InsertProductsJSON()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = app.InsertSalesJSON()
	if err != nil {
		fmt.Println(err)
		return
	}
	// - run
	err = app.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
}
