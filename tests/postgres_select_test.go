package tests_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lauro-ss/goe"
)

func TestPostgresSelect(t *testing.T) {
	db, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}

	err = db.Delete(db.Animal).Where()
	if err != nil {
		t.Fatalf("Expected delete animals, got error: %v", err)
	}
	err = db.Delete(db.Food).Where()
	if err != nil {
		t.Fatalf("Expected delete foods, got error: %v", err)
	}
	err = db.Delete(db.Habitat).Where()
	if err != nil {
		t.Fatalf("Expected delete habitats, got error: %v", err)
	}
	err = db.Delete(db.Info).Where()
	if err != nil {
		t.Fatalf("Expected delete infos, got error: %v", err)
	}
	err = db.Delete(db.Status).Where()
	if err != nil {
		t.Fatalf("Expected delete status, got error: %v", err)
	}
	err = db.Delete(db.Owns).Where()
	if err != nil {
		t.Fatalf("Expected delete owns, got error: %v", err)
	}

	owns := []Owns{
		{Name: "Bones"},
		{Name: "Toys"},
		{Name: "Boxs"},
	}
	err = db.Insert(db.Owns).Value(&owns)
	if err != nil {
		t.Fatalf("Expected insert owns, got error: %v", err)
	}

	weathers := []Weather{
		{Name: "Hot"},
		{Name: "Cold"},
		{Name: "Wind"},
		{Name: "Nice"},
	}
	err = db.Insert(db.Weather).Value(&weathers)
	if err != nil {
		t.Fatalf("Expected insert weathers, got error: %v", err)
	}

	habitats := []Habitat{
		{Id: uuid.New(), Name: "City", IdWeather: weathers[0].Id, NameWeather: "Test"},
		{Id: uuid.New(), Name: "Jungle", IdWeather: weathers[3].Id},
		{Id: uuid.New(), Name: "Savannah", IdWeather: weathers[0].Id},
		{Id: uuid.New(), Name: "Ocean", IdWeather: weathers[2].Id},
	}
	err = db.Insert(db.Habitat).Value(&habitats)
	if err != nil {
		t.Fatalf("Expected insert habitats, got error: %v", err)
	}

	status := []Status{
		{Name: "Cat Alive"},
		{Name: "Dog Alive"},
		{Name: "Big Dog Alive"},
	}

	err = db.Insert(db.Status).Value(&status)
	if err != nil {
		t.Fatalf("Expected insert habitats, got error: %v", err)
	}

	infos := []Info{
		{Id: uuid.New().NodeID(), Name: "Little Cat", IdStatus: status[0].Id, NameStatus: "Test"},
		{Id: uuid.New().NodeID(), Name: "Big Dog", IdStatus: status[2].Id},
	}
	err = db.Insert(db.Info).Value(&infos)
	if err != nil {
		t.Fatalf("Expected insert infos, got error: %v", err)
	}

	animals := []Animal{
		{Name: "Cat", IdHabitat: &habitats[0].Id, IdInfo: &infos[0].Id},
		{Name: "Dog", IdHabitat: &habitats[0].Id, IdInfo: &infos[1].Id},
		{Name: "Forest Cat", IdHabitat: &habitats[1].Id},
		{Name: "Bear", IdHabitat: &habitats[1].Id},
		{Name: "Lion", IdHabitat: &habitats[2].Id},
		{Name: "Puma", IdHabitat: &habitats[1].Id},
		{Name: "Snake", IdHabitat: &habitats[1].Id},
		{Name: "Whale"},
	}
	err = db.Insert(db.Animal).Value(&animals)
	if err != nil {
		t.Fatalf("Expected insert animals, got error: %v", err)
	}

	foods := []Food{{Name: "Meat"}, {Name: "Grass"}}
	err = db.Insert(db.Food).Value(&foods)
	if err != nil {
		t.Fatalf("Expected insert foods, got error: %v", err)
	}

	animalFoods := []int{
		foods[0].Id, animals[0].Id,
		foods[0].Id, animals[1].Id}
	err = db.InsertIn(db.Food, db.Animal).Values(animalFoods)
	if err != nil {
		t.Fatalf("Expected insert animalFoods, got error: %v", err)
	}

	animalOwns := []int{
		animals[0].Id, owns[2].Id,
		animals[1].Id, owns[0].Id,
		animals[1].Id, owns[1].Id,
	}
	err = db.InsertIn(db.Animal, db.Owns).Values(&animalOwns)
	if err != nil {
		t.Fatalf("Expected insert animalOwns, got error: %v", err)
	}

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Select",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(&db.Animal.Id, &db.Animal.Name).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != len(animals) {
					t.Errorf("Expected %v animals, got %v", len(animals), len(a))
				}
			},
		},
		{
			desc: "Select_Where_Equals",
			testCase: func(t *testing.T) {
				var a Animal
				err := db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, animals[0].Id)).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if a.Name != animals[0].Name {
					t.Errorf("Expected a %v, got %v", animals[0].Name, a.Name)
				}
			},
		},
		{
			desc: "Select_Where_Like",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Where(db.Like(&db.Animal.Name, "%Cat%")).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != 2 {
					t.Errorf("Expected %v animals, got %v", 2, len(a))
				}
			},
		},
		{
			desc: "Select_Order_By_Asc",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).OrderByAsc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if a[0].Id > a[1].Id {
					t.Errorf("Expected animals order by asc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Order_By_Desc",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).OrderByDesc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if a[0].Id < a[1].Id {
					t.Errorf("Expected animals order by desc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Page",
			testCase: func(t *testing.T) {
				var a []Animal
				var pageSize uint = 5
				err := db.Select(db.Animal).Page(1, pageSize).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != int(pageSize) {
					t.Errorf("Expected %v animals, got %v", pageSize, len(a))
				}
			},
		},
		{
			desc: "Select_Join",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Food).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != (len(animalFoods) / 2) {
					t.Errorf("Expected 1 animal, got %v", len(a))
				}
				if a[0].Name != animals[0].Name {
					t.Errorf("Expected %v, got %v", animals[0].Name, a[0].Name)
				}
			},
		},
		{
			desc: "Select_Join_Where",
			testCase: func(t *testing.T) {
				var f []Food
				err := db.Select(db.Food).Join(db.Animal, db.Food).Where(
					db.Equals(&db.Animal.Name, animals[0].Name)).Scan(&f)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(f) != 1 {
					t.Errorf("Expected 1 food, got %v", len(f))
				}
				if f[0].Name != foods[0].Name {
					t.Errorf("Expected %v, got %v", foods[0].Name, f[0].Name)
				}
			},
		},
		{
			desc: "Select_Join_Where_And_Equals_Find_0",
			testCase: func(t *testing.T) {
				var f []Food
				err := db.Select(db.Food).Join(db.Animal, db.Food).Where(
					db.Equals(&db.Animal.Name, animals[0].Name),
					db.And(),
					db.Equals(&db.Food.Id, foods[1].Id),
				).Scan(&f)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(f) != 0 {
					t.Errorf("Expected 0 food, got %v", len(f))
				}
			},
		},
		{
			desc: "Select_Inverted_Join_Where_And_Equals_Find_0",
			testCase: func(t *testing.T) {
				var f []Food
				err := db.Select(db.Food).Join(db.Food, db.Animal).Where(
					db.Equals(&db.Animal.Name, animals[0].Name),
					db.And(),
					db.Equals(&db.Food.Id, foods[1].Id),
				).Scan(&f)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(f) != 0 {
					t.Errorf("Expected 0 food, got %v", len(f))
				}
			},
		},
		{
			desc: "Select_Join_Where_And_Equals_Find_1",
			testCase: func(t *testing.T) {
				var f []Food
				err := db.Select(db.Food).Join(db.Animal, db.Food).Where(
					db.Equals(&db.Animal.Name, animals[0].Name),
					db.And(),
					db.Equals(&db.Food.Id, foods[0].Id),
				).Scan(&f)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(f) != 1 {
					t.Errorf("Expected 1 food, got %v", len(f))
				}
			},
		},
		{
			desc: "Select_Join_Order_By_Asc",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Food).OrderByAsc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if a[0].Id > a[1].Id {
					t.Errorf("Expected animals order by asc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Join_Order_By_Desc",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Food).OrderByDesc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if a[0].Id < a[1].Id {
					t.Errorf("Expected animals order by desc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Join_Where_Order_By_Asc",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Food).Where(
					db.Equals(&db.Food.Id, foods[0].Id),
				).OrderByAsc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != 2 {
					t.Errorf("Expected 2 animals, got %v", len(a))
				}
				if a[0].Id > a[1].Id {
					t.Errorf("Expected animals order by asc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Join_Where_Order_By_Desc",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Food).Where(
					db.Equals(&db.Food.Id, foods[0].Id),
				).OrderByDesc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != 2 {
					t.Errorf("Expected 2 animals, got %v", len(a))
				}
				if a[0].Id < a[1].Id {
					t.Errorf("Expected animals order by desc, got %v", a)
				}
			},
		},
		{
			desc: "Select_Join_Many_To_One",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Habitat).OrderByAsc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				for i := range a {
					if a[i].Id != animals[i].Id {
						t.Errorf("Expected %v, got %v", a[0].Id, animals[0].Id)
					}
				}
			},
		},
		{
			desc: "Select_Inverted_Join_Many_To_One",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Habitat, db.Animal).OrderByAsc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				for i := range a {
					if a[i].Id != animals[i].Id {
						t.Errorf("Expected %v, got %v", a[0].Id, animals[0].Id)
					}
				}
			},
		},
		{
			desc: "Select_Left_Join_Many_To_One",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).LeftJoin(db.Habitat, db.Animal).OrderByAsc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != len(animals) {
					t.Errorf("Expected %v, got %v", len(animals), len(a))
				}
				if a[len(a)-1].IdHabitat != nil {
					t.Errorf("Expected nil, got value")
				}
			},
		},
		{
			desc: "Select_Join_Many_To_Many_And_Many_To_One",
			testCase: func(t *testing.T) {
				var f []Food
				err := db.Select(db.Food).Join(db.Food, db.Animal).
					Join(db.Animal, db.Habitat).Where(db.Equals(&db.Habitat.Id, habitats[0].Id)).
					Scan(&f)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(f) != 2 {
					t.Errorf("Expected 2, got : %v", len(f))
				}
			},
		},
		{
			desc: "Select_Join_One_To_One",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Info).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != 2 {
					t.Errorf("Expected 2, got : %v", len(a))
				}
			},
		},
		{
			desc: "Select_Inverted_Join_One_To_One",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Info, db.Animal).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != 2 {
					t.Errorf("Expected 2, got : %v", len(a))
				}
			},
		},
		{
			desc: "Select_Animal_Join_One_To_One_And_Many_To_Many",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Info).Join(db.Animal, db.Food).
					Where(db.Equals(&db.Food.Id, foods[0].Id)).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != 2 {
					t.Errorf("Expected 2, got : %v", len(a))
				}
			},
		},
		{
			desc: "Select_Info_Join_One_To_One_And_Many_To_Many",
			testCase: func(t *testing.T) {
				var i []Info
				err := db.Select(db.Info).Join(db.Animal, db.Info).Join(db.Animal, db.Food).
					Where(db.Equals(&db.Food.Id, foods[0].Id)).Scan(&i)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(i) != 2 {
					t.Errorf("Expected 2, got : %v", len(i))
				}
			},
		},
		{
			desc: "Select_Info_Inverted_Join_One_To_One_And_Many_To_Many",
			testCase: func(t *testing.T) {
				var i []Info
				err := db.Select(db.Info).Join(db.Info, db.Animal).Join(db.Animal, db.Food).
					Where(db.Equals(&db.Food.Id, foods[0].Id)).Scan(&i)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(i) != 2 {
					t.Errorf("Expected 2, got : %v", len(i))
				}
			},
		},
		{
			desc: "Select_Info_Join_Status_One_To_One_And_Many_To_Many",
			testCase: func(t *testing.T) {
				var s []Info
				err := db.Select(db.Info).Join(db.Status, db.Info).Join(db.Animal, db.Info).
					Join(db.Animal, db.Food).Where(db.Equals(&db.Food.Id, foods[0].Id)).Scan(&s)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(s) != 2 {
					t.Errorf("Expected 2, got : %v", len(s))
				}
			},
		},
		{
			desc: "Select_Status_Inverted_Join_One_To_One_And_Many_To_Many",
			testCase: func(t *testing.T) {
				var s []Status
				err := db.Select(db.Status).Join(db.Info, db.Status).Join(db.Info, db.Animal).
					Join(db.Animal, db.Food).Where(db.Equals(&db.Food.Id, foods[0].Id)).Scan(&s)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(s) != 2 {
					t.Errorf("Expected 2, got : %v", len(s))
				}
			},
		},
		{
			desc: "Select_Status_Join_One_To_One_And_Many_To_Many",
			testCase: func(t *testing.T) {
				var s []Status
				err := db.Select(db.Status).Join(db.Status, db.Info).Join(db.Animal, db.Info).
					Join(db.Food, db.Animal).Where(db.Equals(&db.Food.Id, foods[0].Id)).Scan(&s)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(s) != 2 {
					t.Errorf("Expected 2, got : %v", len(s))
				}
			},
		},
		{
			desc: "Select_Animal_Join_Many_To_Many_And_Many_To_Many",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Food).Join(db.Animal, db.Owns).
					Where(db.Equals(&db.Owns.Id, owns[0].Id), db.And(), db.Equals(&db.Food.Id, foods[0].Id)).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != 1 {
					t.Errorf("Expected 2, got : %v", len(a))
				}
				if a[0].Id != animals[1].Id {
					t.Errorf("Expected Dog, got : %v", a)
				}
			},
		},
		{
			desc: "Select_Animal_Inverted_Join_Many_To_Many_And_Many_To_Many",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Food, db.Animal).Join(db.Owns, db.Animal).
					Where(db.Equals(&db.Owns.Id, owns[0].Id), db.And(), db.Equals(&db.Food.Id, foods[0].Id)).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != 1 {
					t.Errorf("Expected 2, got : %v", len(a))
				}
				if a[0].Id != animals[1].Id {
					t.Errorf("Expected Dog, got : %v", a)
				}
			},
		},
		{
			desc: "Select_Animal_By_Weather_Join_One_To_Many",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Habitat).Join(db.Habitat, db.Weather).
					Where(db.Equals(&db.Weather.Id, weathers[3].Id)).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(a) != 4 {
					t.Errorf("Expected 4, got : %v", len(a))
				}
			},
		},
		{
			desc: "Select_Weather_By_Animal_Join_One_To_Many",
			testCase: func(t *testing.T) {
				var w []Weather
				err := db.Select(db.Weather).Join(db.Weather, db.Habitat).Join(db.Habitat, db.Animal).
					Where(db.Equals(&db.Animal.Id, animals[0].Id)).Scan(&w)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(w) != 1 {
					t.Errorf("Expected 1, got : %v", len(w))
				}
			},
		},
		{
			desc: "Select_Join_Page",
			testCase: func(t *testing.T) {
				var a []Animal
				var pageSize uint = 2
				err := db.Select(db.Animal).Join(db.Animal, db.Food).Page(1, pageSize).Scan(&a)
				if err != nil {
					t.Errorf("Expected a page select, got error: %v", err)
				}
				if len(a) != int(pageSize) {
					t.Errorf("Expected %v animals, got %v", pageSize, len(a))
				}
			},
		},
		{
			desc: "Select_Anonymous_Struct",
			testCase: func(t *testing.T) {
				var a struct {
					Id1 int
					Id2 string
				}
				err := db.Select(&db.Animal.Id, &db.Animal.Name).OrderByAsc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if a.Id1 != animals[0].Id {
					t.Errorf("Expected %v, got : %v", animals[0].Id, a.Id1)
				}
				if a.Id2 != animals[0].Name {
					t.Errorf("Expected %v, got : %v", animals[0].Name, a.Id2)
				}
			},
		},
		{
			desc: "Select_Anonymous_Struct_2",
			testCase: func(t *testing.T) {
				var a struct {
					Id  int
					Id2 string
				}
				err := db.Select(&db.Animal.Id, &db.Animal.Name).OrderByAsc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if a.Id != animals[0].Id {
					t.Errorf("Expected %v, got : %v", animals[0].Id, a.Id)
				}
				if a.Id2 != animals[0].Name {
					t.Errorf("Expected %v, got : %v", animals[0].Name, a.Id2)
				}
			},
		},
		{
			desc: "Select_Anonymous_Struct_Slice_3",
			testCase: func(t *testing.T) {
				var a []struct {
					AnimalId        int
					AnimalName      string
					AnimalIdHabitat uuid.UUID
					AnimalIdInfo    []byte
					HabitatId       uuid.UUID
					HabitatName     string
					IdWeather       int
					NameWeather     string
				}
				err := db.Select(db.Animal, db.Habitat).Join(db.Animal, db.Habitat).OrderByAsc(&db.Animal.Id).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				for i := range a {
					if a[i].AnimalId != animals[i].Id {
						t.Errorf("Expected %v, got %v", a[0].AnimalId, animals[0].Id)
					}
					if a[i].AnimalIdHabitat.String() != a[i].HabitatId.String() {
						t.Errorf("Expected %v, got %v", a[i].AnimalIdHabitat.String(), a[i].HabitatId.String())
					}
				}
			},
		},
		{
			desc: "Select_Invalid_Scan",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Scan(a)
				if !errors.Is(err, goe.ErrInvalidScan) {
					t.Errorf("Expected goe.ErrInvalidScan, got error: %v", err)
				}
			},
		},
		{
			desc: "Select_Invalid_OrderBy",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).OrderByAsc(db.Animal.IdHabitat).Scan(&a)
				if !errors.Is(err, goe.ErrInvalidOrderBy) {
					t.Errorf("Expected goe.ErrInvalidOrderBy, got error: %v", err)
				}
			},
		},
		{
			desc: "Select_Invalid_Where",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Where(db.Equals(db.Animal.Id, 1)).Scan(&a)
				if !errors.Is(err, goe.ErrInvalidWhere) {
					t.Errorf("Expected goe.ErrInvalidWhere, got error: %v", err)
				}
			},
		},
		{
			desc: "Select_Invalid_Join",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(db.Animal).Join(db.Animal, db.Weather).Scan(&a)
				if !errors.Is(err, goe.ErrNoMatchesTables) {
					t.Errorf("Expected goe.ErrNoMatchesTables, got error: %v", err)
				}
			},
		},
		{
			desc: "Select_Invalid_Arg",
			testCase: func(t *testing.T) {
				var a []Animal
				err := db.Select(nil).Join(db.Animal, db.Weather).Scan(&a)
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected goe.ErrInvalidArg, got error: %v", err)
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
