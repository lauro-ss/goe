package main

import (
	"github.com/lauro-ss/goe"
)

type Product struct {
	Id   int
	Name string
}

type Categoria struct {
	Id   int
	Name string
}

type ProdCat struct {
	Id int `goe:"map=produto.id"`
}

func main() {
	db := goe.Connect()
	// "db.Get(&users).Join('Categoria')"
	// "db.Select(&user)"
	// "db.Select('Id','Name', '')"
	db.Select("Produto.Id", "Produto.Name", "Categoria.Name", "Subcategoria.Name").
		From("Produto").Join("Categoria").Join("Subcategoria")
}
