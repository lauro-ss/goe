package main

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lauro-ss/goe"
)

type Produto struct {
	Id         string `goe:"pk;t:uuid"`
	Name       string `goe:"t:varchar(20)"`
	Categorias []Categoria
}

type Categoria struct {
	Id            string `goe:"pk;t:uuid"`
	Name          string `goe:"t:varchar(20)"`
	Produtos      []Produto
	Subcategorias []Subcategoria
}

type Subcategoria struct {
	Id         string `goe:"pk;t:uuid"`
	Name       string `goe:"t:varchar(20)"`
	Categorias []Categoria
}

type Animal struct {
	Id    string
	Name  string
	Emoji string
	Foods []Food
}

type Food struct {
	Id      string
	Name    string
	Emoji   string
	Animals []Animal
}

type AnimalDb struct {
	Id    *goe.Pk
	Name  *goe.Att
	Emoji *goe.Att
}

type FoodDb struct {
	Id    *goe.Pk
	Name  *goe.Att
	Emoji *goe.Att
}

// TODO: Check if field exists
type Database struct {
	Animal AnimalDb
	Food   FoodDb
	*goe.DB
}

func main() {

	// db := goe.Connect("database_conection", goe.Config{MigrationsPath: "./Migrations"})
	// db.Migrate(&Produto{})
	// db.Migrate(&Categoria{})
	// db.Migrate(&Subcategoria{})
	// db := &Database{
	// 	Animal: AnimalDb{
	// 		Id: goe.MapAttribute(&Animal{}, "Id"),
	// 	},
	// }

	db := &Database{DB: &goe.DB{}}
	//goe.Map(db.Animal, &Animal{})
	//goe.Connect(db)
	goe.Map(db, Animal{})
	goe.Map(db, Food{})
	// err := goe.Map(&db.Animal, Animal{})
	// fmt.Println(err)
	fmt.Printf("%p \n", db.Animal.Id.Fk["Food"])
	fmt.Printf("%p \n", db.Food.Id)

	fmt.Println("Next")

	fmt.Printf("%p \n", db.Food.Id.Fk["Animal"])
	fmt.Printf("%p \n", db.Animal.Id)
	//"db.Select(&users).Where(user.Id.Equals(1).Or())"
	db.Open("pgx", "user=app password=123456 host=localhost port=5432 database=appanimal sslmode=disable")

	// ids := make([]string, 10)

	//works
	// var ids []string
	// db.Select(db.Animal.Id).Result(&ids)
	// fmt.Println(db.Errors())
	// fmt.Println(ids)

	// var animals []Animal
	// db.Result(&animals)
	// fmt.Println(db.Erros)
	// fmt.Println(animals)

	//db.Select(db.Animal.Id).Where(db.Animal.Id.Equals("1"))

	// db.SetTable(&Produto{})
	// db.SetTable(&Categoria{})
	// "db.Get(&users).Join('Categoria')"
	// "db.Select(&user)"
	// "db.Select('Id','Name', '')"
}
