package main

import (
	"github.com/lauro-ss/goe"
)

func main() {
	db := goe.Connect()
	db.Select()
}
