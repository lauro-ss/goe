package tests_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/olauro/goe"
)

func TestPostgresDelete(t *testing.T) {
	db, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}
	if db.ConnPool.Stats().InUse != 0 {
		t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
	}

	err = db.DeleteIn(db.Animal, db.Food).Where()
	if err != nil {
		t.Fatalf("Expected delete AnimalFood, got error: %v", err)
	}
	err = db.DeleteIn(db.Animal, db.Owns).Where()
	if err != nil {
		t.Fatalf("Expected delete AnimalOwns, got error: %v", err)
	}
	err = db.Delete(db.Flag).Where()
	if err != nil {
		t.Fatalf("Expected delete flags, got error: %v", err)
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
	err = db.Delete(db.UserRole).Where()
	if err != nil {
		t.Fatalf("Expected delete user roles, got error: %v", err)
	}
	err = db.Delete(db.User).Where()
	if err != nil {
		t.Fatalf("Expected delete users, got error: %v", err)
	}
	err = db.Delete(db.Role).Where()
	if err != nil {
		t.Fatalf("Expected delete roles, got error: %v", err)
	}

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Delete_One_Record",
			testCase: func(t *testing.T) {
				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				a := Animal{Name: "Dog"}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				var as Animal
				err = db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&as)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				err = db.Delete(db.Animal).Where(db.Equals(&db.Animal.Id, as.Id))
				if err != nil {
					t.Errorf("Expected a delete animal, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				err = db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&as)
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}

				var id int
				err = db.Select(&db.Animal.Id).Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&id)
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}
			},
		},
		{
			desc: "Delete_All_Records",
			testCase: func(t *testing.T) {
				animals := []Animal{
					{Name: "Cat"},
					{Name: "Forest Cat"},
					{Name: "Catt"},
				}
				err = db.Insert(db.Animal).Value(&animals)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}

				animals = nil
				err = db.Select(db.Animal).Where(db.Like(&db.Animal.Name, "%Cat%")).Scan(&animals)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}

				var a Animal
				err = db.Select(db.Animal).Where(db.Like(&db.Animal.Name, "%Cat%")).Scan(&a)
				if err != nil {
					t.Errorf("Expected a select one animal, got error: %v", err)
				}

				err = db.Delete(db.Animal).Where(db.Like(&db.Animal.Name, "%Cat%"))
				if err != nil {
					t.Errorf("Expected a delete, got error: %v", err)
				}

				animals = nil
				err = db.Select(db.Animal).Where(db.Like(&db.Animal.Name, "%Cat%")).Scan(&animals)
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}

				if len(animals) != 0 {
					t.Errorf(`Expected to delete all "Cat" animals, got: %v`, len(animals))
				}
			},
		},
		{
			desc: "DeleteIn_One_Record",
			testCase: func(t *testing.T) {
				animals := Animal{Name: "Cat"}
				err = db.Insert(db.Animal).Value(&animals)
				if err != nil {
					t.Errorf("Expected insert animals, got error: %v", err)
				}

				foods := Food{Id: uuid.New(), Name: "Meat"}
				err = db.Insert(db.Food).Value(&foods)
				if err != nil {
					t.Errorf("Expected insert foods, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				err = db.InsertIn(db.Animal, db.Food).Values(animals.Id, foods.Id)
				if err != nil {
					t.Errorf("Expected insert animalFood, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				var AnimalFood *struct {
					AnimalName string
					FoodName   string
				}
				err = db.Select(&db.Animal.Name, &db.Food.Name).
					Join(db.Animal, db.Food).Where(
					db.Equals(&db.Animal.Id, animals.Id),
					db.And(),
					db.Equals(&db.Food.Id, foods.Id)).Scan(&AnimalFood)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}

				if AnimalFood.AnimalName != animals.Name || AnimalFood.FoodName != foods.Name {
					t.Errorf(`Expected %v got %v, Expected %v got: %v`,
						animals.Name, AnimalFood.AnimalName,
						foods.Name, AnimalFood.FoodName,
					)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				err = db.DeleteIn(db.Animal, db.Food).Where(
					db.Equals(&db.Animal.Id, animals.Id),
					db.And(),
					db.Equals(&db.Food.Id, foods.Id))
				if err != nil {
					t.Errorf("Expected a delete animalFood, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				err = db.Select(&db.Animal.Name, &db.Food.Name).
					Join(db.Animal, db.Food).Where(
					db.Equals(&db.Animal.Id, animals.Id),
					db.And(),
					db.Equals(&db.Food.Id, foods.Id)).Scan(&AnimalFood)
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}
			},
		},
		{
			desc: "DeleteIn_All_Record",
			testCase: func(t *testing.T) {
				animals := Animal{Name: "Cat"}
				err = db.Insert(db.Animal).Value(&animals)
				if err != nil {
					t.Errorf("Expected insert animals, got error: %v", err)
				}

				foods := []Food{
					{Id: uuid.New(), Name: "Meat"},
					{Id: uuid.New(), Name: "Grass"},
				}
				err = db.Insert(db.Food).Value(&foods)
				if err != nil {
					t.Errorf("Expected insert foods, got error: %v", err)
				}

				animalFods := []any{
					animals.Id, foods[0].Id,
					animals.Id, foods[1].Id,
				}
				err = db.InsertIn(db.Animal, db.Food).Values(animalFods)
				if err != nil {
					t.Errorf("Expected insert animalFood, got error: %v", err)
				}

				var AnimalFood []struct {
					AnimalName string
					FoodName   string
				}
				err = db.Select(&db.Animal.Name, &db.Food.Name).
					Join(db.Animal, db.Food).Where(
					db.Equals(&db.Animal.Id, animals.Id)).Scan(&AnimalFood)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}
				if len(AnimalFood) != 2 {
					t.Errorf("Expected a 2, got: %v", len(AnimalFood))
				}

				err = db.DeleteIn(db.Animal, db.Food).Where(
					db.Equals(&db.Animal.Id, animals.Id))
				if err != nil {
					t.Errorf("Expected a delete animalFood, got error: %v", err)
				}

				err = db.Select(&db.Animal.Name, &db.Food.Name).
					Join(db.Animal, db.Food).Where(
					db.Equals(&db.Animal.Id, animals.Id)).Scan(&AnimalFood)
				if !errors.Is(err, goe.ErrNotFound) {
					t.Errorf("Expected a select, got error: %v", err)
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
