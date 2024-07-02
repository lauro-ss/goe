package main

import (
	"fmt"
	"time"

	"github.com/lauro-ss/goe"
	"github.com/lauro-ss/goe/postgres"
)

type Flag struct {
	Id       []byte
	Value32  float32
	Value    float64
	CreateAt time.Time
	Ative    bool
}

type Animal struct {
	Id       string  `goe:"pk;type:uuid"`
	Emoji    string  `goe:"index(n:idx_emoji)"`
	Name     string  `goe:"type:varchar(30);index(n:idx_name_low f:lower, n:idx_name_tail f:upper)"`
	Tails    *string `goe:"index(n:idx_name_tail f:lower)"`
	Foods    []Food  `goe:"table:AnimalFood"`
	Status   []Status
	Habitats []Habitat `goe:"table:AnimalHabitat"`
}

type Habitat struct {
	Id       int
	Name     string
	Bits     []byte
	Weathers []Weather
	Animals  []Animal `goe:"table:AnimalHabitat"`
	Foods    []Food   `goe:"table:FoodHabitat"`
}

type Status struct {
	Id    string `goe:"type:uuid"`
	Name  string
	Alive bool
	Animal
}

type Food struct {
	Id       string `goe:"type:uuid"`
	Name     string
	Animals  []Animal  `goe:"table:AnimalFood"`
	Habitats []Habitat `goe:"table:FoodHabitat"`
	Emoji    string
}

type Weather struct {
	Name  string
	Id    int
	Emoji string
	*Habitat
}

// TODO: Check if field exists
type Database struct {
	Animal  *Animal
	Food    *Food
	Status  *Status
	Habitat *Habitat
	Weather *Weather
	Flag    *Flag
	*goe.DB
}

func main() {
	db := &Database{DB: &goe.DB{}}
	goe.Open(db, postgres.Open("user=app password=123456 host=localhost port=5432 database=orm sslmode=disable"))
	db.Migrate(goe.MigrateFrom(db))

	start := time.Now()
	TestIn(db)
	fmt.Println(time.Since(start))
}

func TestIn(db *Database) {
	t := "tails"
	db.Insert(db.Animal).Value(&Animal{Id: "5ad0e5fc-e9f7-4855-9698-d0c10b996f73", Name: "Lion", Emoji: "🦁", Tails: &t})
	db.Insert(db.Animal).Value(&Animal{Id: "906f4f1f-49e7-47ee-8954-2d6e0a3354cf", Name: "Cat", Emoji: "🐱"})

	db.Insert(db.Food).Value(&Food{Id: "401b5e23-5aa7-435e-ba4d-5c1b2f123596", Name: "Meat", Emoji: "🥩"})
	db.Insert(db.Food).Value(&Food{Id: "f023a4e7-34e9-4db2-85e0-efe8d67eea1b", Name: "Hotdog", Emoji: "🌭"})
	db.Insert(db.Food).Value(&Food{Id: "fc1865b4-6f2d-4cc6-b766-49c2634bf5c4", Name: "Cookie", Emoji: "🍪"})

	db.InsertIn(db.Animal, db.Food).Values("5ad0e5fc-e9f7-4855-9698-d0c10b996f73", "401b5e23-5aa7-435e-ba4d-5c1b2f123596") // 🦁 eats 🥩
	db.InsertIn(db.Animal, db.Food).Values("906f4f1f-49e7-47ee-8954-2d6e0a3354cf", "401b5e23-5aa7-435e-ba4d-5c1b2f123596") // 🐱 eats 🥩
	db.InsertIn(db.Food, db.Animal).Values("f023a4e7-34e9-4db2-85e0-efe8d67eea1b", "906f4f1f-49e7-47ee-8954-2d6e0a3354cf") // 🐱 eats 🌭

	var a []Animal
	db.Select(db.Animal).Where(db.Equals(&db.Food.Id, "401b5e23-5aa7-435e-ba4d-5c1b2f123596")).Result(&a)
	fmt.Println(a)

	var aStruct Animal
	db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, "906f4f1f-49e7-47ee-8954-2d6e0a3354cf")).Result(&aStruct)

	t = "teste_2"
	aStruct.Tails = &t
	db.Update(db.Animal).Where(db.Equals(&db.Animal.Id, aStruct.Id)).Value(aStruct)

	db.Select(db.Animal).Where(db.Equals(&db.Food.Id, "401b5e23-5aa7-435e-ba4d-5c1b2f123596")).Result(&a)
	fmt.Println(a)

	s := Status{
		Id:     "8af13eac-abe6-4582-8566-6f15ad55e5cf",
		Name:   "Eating",
		Alive:  true,
		Animal: aStruct,
	}

	db.Insert(db.Status).Value(&s)

	var f []Food
	//db.Select(db.Food).Join(db.Animal).Where(db.Equals(&db.Status.Id, "401b5e23-5aa7-435e-ba4d-5c1b2f123596")).Result(&f)
	db.Select(db.Food).Join(db.Animal).Where(db.Equals(&db.Status.Alive, true)).Result(&f)
	fmt.Println(f)

	db.Insert(db.Flag).Value(&Flag{Id: []byte{0, 1}, Value32: 1.1, Value: 2.1, CreateAt: time.Now(), Ative: true})

	var flags []Flag
	db.Select(db.Flag).Result(&flags)
	fmt.Println(flags)

	var id []string
	db.Select(&db.Status.Animal).Result(&id)
	fmt.Println(id)

	db.UpdateIn(db.Animal, db.Food).Where(
		db.Equals(&db.Animal.Id, "906f4f1f-49e7-47ee-8954-2d6e0a3354cf"),
		db.And(),
		db.Equals(&db.Food.Id, "f023a4e7-34e9-4db2-85e0-efe8d67eea1b")).
		Value("fc1865b4-6f2d-4cc6-b766-49c2634bf5c4")

	db.Select(db.Food).Join(db.Animal).Where(db.Equals(&db.Status.Alive, true)).Result(&f)
	fmt.Println(f)

	db.DeleteIn(db.Animal, db.Food).Where(db.Equals(&db.Food.Id, "fc1865b4-6f2d-4cc6-b766-49c2634bf5c4"))

	db.Select(db.Food).Join(db.Animal).Where(db.Equals(&db.Status.Alive, true)).Result(&f)
	fmt.Println(f)

	db.DeleteIn(db.Animal, db.Food).
		Where(
			db.Equals(&db.Food.Id, "401b5e23-5aa7-435e-ba4d-5c1b2f123596"),
			db.And(),
			db.Equals(&db.Animal.Id, "906f4f1f-49e7-47ee-8954-2d6e0a3354cf"),
		)

	db.Select(db.Food).Join(db.Animal).Where(db.Equals(&db.Status.Alive, true)).Result(&f)
	fmt.Println(f)

	DeleteAll(db)
}

func TestInsert(db *Database, i int) {
	for c := 0; c < i; c++ {
		db.Insert(db.Habitat).Value(&Habitat{Name: "Test", Bits: []byte{0, 3}})
	}
}

func DeleteAll(db *Database) {
	db.DeleteIn(db.Animal, db.Food).Where()
	db.DeleteIn(db.Animal, db.Habitat).Where()
	db.DeleteIn(db.Food, db.Habitat).Where()
	db.Delete(db.Status).Where()
	db.Delete(db.Animal).Where()
	db.Delete(db.Food).Where()
	db.Delete(db.Weather).Where()
	db.Delete(db.Habitat).Where()
	db.Delete(db.Flag).Where()
}
