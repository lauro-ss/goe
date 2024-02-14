package main

import (
	"github.com/lauro-ss/goe"
)

type Produto struct {
	Id         string `goe:"pk;t:uuid"`
	Name       string `goe:"t:varchar(20)"`
	Categorias []Categoria
}

type Categoria struct {
	Id       string `goe:"pk;t:uuid"`
	Name     string `goe:"t:varchar(20)"`
	Produtos []Produto
}

func main() {

	db := goe.Connect("database_conection", goe.Config{MigrationsPath: "./Migrations"})
	db.Migrate(&Produto{})
	db.Migrate(&Categoria{})
	// db.SetTable(&Produto{})
	// db.SetTable(&Categoria{})
	// "db.Get(&users).Join('Categoria')"
	// "db.Select(&user)"
	// "db.Select('Id','Name', '')"
}
