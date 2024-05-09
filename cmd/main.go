package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lauro-ss/goe"
)

// type Produto struct {
// 	Id         string `goe:"pk;t:uuid"`
// 	Name       string `goe:"t:varchar(20)"`
// 	Categorias []Categoria
// }

// type Categoria struct {
// 	Id            string `goe:"pk;t:uuid"`
// 	Name          string `goe:"t:varchar(20)"`
// 	Produtos      []Produto
// 	Subcategorias []Subcategoria
// }

// type Subcategoria struct {
// 	Id         string `goe:"pk;t:uuid"`
// 	Name       string `goe:"t:varchar(20)"`
// 	Categorias []Categoria
// }

type Animal struct {
	Id       string `goe:"pk"`
	Emoji    string
	Name     string
	Foods    []Food `goe:"table:AnimalFood"`
	Status   []Status
	Habitats []Habitat `goe:"table:AnimalHabitat"`
}

type Habitat struct {
	Id      int
	Name    string
	Animals []Animal `goe:"table:AnimalHabitat"`
}

type Status struct {
	Id    string
	Name  string
	Alive bool
	Animal
}

type Food struct {
	Id      string `goe:"pk;t:uuid"`
	Name    string
	Animals []Animal `goe:"table:AnimalFood"`
	Emoji   string
}

// type AnimalDb struct {
// 	Id    goe.Pk
// 	Name  goe.Att
// 	Emoji goe.Att
// }

// type StatusDb struct {
// 	Id   goe.Pk
// 	Name goe.Att
// }

// type AnimalFood struct {
// 	IdAnimal string
// 	IdFood   string
// }

// type FoodDb struct {
// 	IdFood goe.Pk
// 	Name   goe.Att
// 	Emoji  goe.Att
// }

// TODO: Check if field exists
type Database struct {
	Animal  *Animal
	Food    *Food
	Status  *Status
	Habitat *Habitat
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
	goe.Open(db, "pgx", "user=app password=123456 host=localhost port=5432 database=appanimal sslmode=disable")
	//goe.Map(db.Animal, &Animal{})
	//goe.Connect(db)
	// goe.Map(db, Status{})
	// goe.Map(db, Animal{})
	// goe.Map(db, Food{})

	// fmt.Println(db.Animal.IdAnimal)
	// fmt.Printf("%p \n", db.Animal.IdAnimal)
	// fmt.Println(db.Status.Id)
	// fmt.Println(db.Status.Name)
	// fmt.Printf("%p \n", db.Status.Id)
	// err := goe.Map(&db.Animal, Animal{})
	// fmt.Println(err)
	//fmt.Printf("%p \n", db.Animal.IdAnimal.Fk["Food"])
	// fmt.Printf("%p Food \n", db.Food.IdFood)
	// // fmt.Println(db.Animal.Name)
	// fmt.Println("Next")
	// fmt.Println(db.AnimalFood.IdAnimal)
	// fmt.Println(db.Animal.IdAnimal)
	// fmt.Printf("%p, %p \n", db.Animal.IdAnimal, db.AnimalFood.IdAnimal)
	// fmt.Printf("%p \n", db.AnimalFood.IdFood)
	// fmt.Println(db.AnimalFood.IdFood)
	// fmt.Println(db.Food.IdFood)
	// fmt.Printf("%p \n", db.Animal.IdAnimal)
	// fmt.Println(db.AnimalFood.IdFood)
	// fmt.Println(db.AnimalFood.IdAnimal)
	// fmt.Printf("%p \n", db.Food.IdFood)
	// fmt.Println(db.Food.IdFood)
	// fmt.Println(db.Animal.IdAnimal)
	// fmt.Printf("%p \n", db.AnimalFood.IdAnimal)
	// fmt.Printf("%p \n", db.AnimalFood.IdFood)
	// fmt.Printf("%p Animal \n", db.Animal.IdAnimal)
	// CheckManyToOne(db)
	// CheckManyToMany(db)
	// fmt.Println(db.Animal.Emoji, db.Food.Emoji)
	//db.Select(db.Animal.IdAnimal)

	//db.Open("pgx", "user=app password=123456 host=localhost port=5432 database=appanimal sslmode=disable")

	// go func() {
	// 	a := make([]Animal, 10)
	// 	db.Select(db.Food.IdFood, db.Animal.Emoji).Result(&a)
	// 	fmt.Println(a)
	// }()

	// go func() {
	// 	a := make([]Animal, 10)
	// 	db.Select(db.Food.IdFood, db.Animal.Emoji).Result(&a)
	// 	fmt.Println(a)
	// }()

	//fmt.Println(db.Animal)
	// var t []struct {
	// 	Id string
	// }

	// 	db.Select(&db.Food.Id).Result(&t)
	// 	fmt.Println(t)

	// for i := 0; i < 100; i++ {
	// 	go func() {
	// 		a := make([]Animal, 0)
	// 		db.Select(db.Animal).Where(db.Equals(&db.Food.Id, "ae5bf981-788c-46c0-aa4d-66dc632fbe47")).Result(&a)
	// 		fmt.Println(a)
	// 	}()
	// }
	// time.Sleep(3 * time.Second)
	// a := make([]Animal, 0)
	// db.Select(db.Animal, db.Status).Where(db.Equals(&db.Food.Id, "ae5bf981-788c-46c0-aa4d-66dc632fbe47")).Result(&a)
	// fmt.Println(a)
	// animal := Animal{
	// 	Id:    "8583db14-7ea7-4912-9b0c-ba33700c1e09",
	// 	Name:  "Cow",
	// 	Emoji: "Emoji",
	// }
	// db.Insert(db.Animal).Result(&animal)
	// fmt.Println(animal)

	// h := make([]Habitat, 0)
	// db.Select(db.Habitat).Result(&h)
	// fmt.Println(h)
	// h := &Habitat{Name: "Floresta"}
	// db.Insert(db.Habitat).Value(h)
	// fmt.Println(h)
	// v := struct {
	// 	IdAnimal string
	// 	IdFood   string
	// }{
	// 	IdAnimal: "",
	// 	IdFood:   "",
	// }
	//db.InsertBetwent(db.Animal, db.Food).Result("408834cc-bbdf-4173-bcae-34aaacfcd5fe", "523da8fd-3e75-4220-a244-a2a73a21ae3e")
	h := &Habitat{Name: "Floresta"}
	db.Update(db.Habitat).Where(db.Equals(&db.Habitat.Id, "teste1")).Result(h)
	// db.Select(db.Status).Where(db.Equals(&db.Status.Alive, false)).Result(&a)

	// db.Select(db.Food.Name).Result(nil)
	// ids := make([]string, 10)
	//works
	// var ids []string
	// db.Equals(db.Animal.Id, 1)
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

// func CheckManyToOne(db *Database) {
// 	ap := fmt.Sprintf("%p", db.Animal.IdAnimal)
// 	sp := fmt.Sprintf("%p", db.Status.Id)

// 	s := fmt.Sprint(db.Status.Id)
// 	a := fmt.Sprint(db.Animal.IdAnimal)

// 	if !strings.Contains(s, ap) {
// 		fmt.Println("Fail on " + ap + " " + s)
// 	}

// 	if !strings.Contains(a, sp) {
// 		fmt.Println("Fail on " + sp + " " + a)
// 	}
// }

// func CheckManyToMany(db *Database) {
// 	ap := fmt.Sprintf("%p", db.Animal.IdAnimal)
// 	fp := fmt.Sprintf("%p", db.Food.IdFood)
// 	f := fmt.Sprint(db.AnimalFood.IdFood)
// 	a := fmt.Sprint(db.AnimalFood.IdAnimal)
// 	if !strings.Contains(f, ap) && !strings.Contains(a, fp) {
// 		fmt.Println("Fail on " + ap + " " + f)
// 		fmt.Println("Fail on " + fp + " " + a)
// 	}

// 	ap = fmt.Sprintf("%p", db.AnimalFood.IdAnimal)
// 	fp = fmt.Sprintf("%p", db.AnimalFood.IdFood)
// 	f = fmt.Sprint(db.Food.IdFood)
// 	a = fmt.Sprint(db.Animal.IdAnimal)
// 	if !strings.Contains(f, fp) && !strings.Contains(a, ap) {
// 		fmt.Println("Fail pointer " + ap + " on " + a)
// 		fmt.Println("Fail pointer " + fp + " on " + f)
// 	}
// }
