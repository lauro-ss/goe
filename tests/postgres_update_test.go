package tests_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olauro/goe"
)

func TestPostgresUpdate(t *testing.T) {
	db, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Update_Flag",
			testCase: func(t *testing.T) {
				f := Flag{
					Id:      uuid.New(),
					Name:    "Flag",
					Float32: 1.1,
					Float64: 2.2,
					Today:   time.Now(),
					Int:     -1,
					Int8:    -8,
					Int16:   -16,
					Int32:   -32,
					Int64:   -64,
					Uint:    1,
					Uint8:   8,
					Uint16:  16,
					Uint32:  32,
					Bool:    true,
				}
				err = db.Insert(db.Flag).Value(&f)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}

				ff := Flag{
					Name:    "Flag_Test",
					Float32: 3.3,
					Float64: 4.4,
					Bool:    false,
				}
				err = db.Update(&db.Flag.Name, &db.Flag.Bool, &db.Flag.Float64, &db.Flag.Float32).Where(db.Equals(&db.Flag.Id, f.Id)).Value(ff)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				var fselect Flag
				err = db.Select(db.Flag).Where(db.Equals(&db.Flag.Id, f.Id)).Scan(&fselect)

				if fselect.Name != ff.Name {
					t.Errorf("Expected a update on name, got : %v", fselect.Name)
				}
				if fselect.Float32 != ff.Float32 {
					t.Errorf("Expected a update on float32, got : %v", fselect.Float32)
				}
				if fselect.Float64 != ff.Float64 {
					t.Errorf("Expected a update on float32, got : %v", fselect.Float64)
				}
				if fselect.Bool != ff.Bool {
					t.Errorf("Expected a update on float32, got : %v", fselect.Bool)
				}
			},
		},
		{
			desc: "Update_Flag_Pointer",
			testCase: func(t *testing.T) {
				f := Flag{
					Id:      uuid.New(),
					Name:    "Flag",
					Float32: 1.1,
					Float64: 2.2,
					Today:   time.Now(),
					Int:     -1,
					Int8:    -8,
					Int16:   -16,
					Int32:   -32,
					Int64:   -64,
					Uint:    1,
					Uint8:   8,
					Uint16:  16,
					Uint32:  32,
					Bool:    true,
				}
				err = db.Insert(db.Flag).Value(&f)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}

				ff := Flag{
					Name:    "Flag_Test",
					Float32: 3.3,
					Float64: 4.4,
					Bool:    false,
				}
				err = db.Update(&db.Flag.Name, &db.Flag.Bool, &db.Flag.Float64, &db.Flag.Float32).Where(db.Equals(&db.Flag.Id, f.Id)).Value(&ff)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				var fselect Flag
				err = db.Select(db.Flag).Where(db.Equals(&db.Flag.Id, f.Id)).Scan(&fselect)

				if fselect.Name != ff.Name {
					t.Errorf("Expected a update on name, got : %v", fselect.Name)
				}
				if fselect.Float32 != ff.Float32 {
					t.Errorf("Expected a update on float32, got : %v", fselect.Float32)
				}
				if fselect.Float64 != ff.Float64 {
					t.Errorf("Expected a update on float32, got : %v", fselect.Float64)
				}
				if fselect.Bool != ff.Bool {
					t.Errorf("Expected a update on float32, got : %v", fselect.Bool)
				}
			},
		},
		{
			desc: "Update_Animal",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				w := Weather{
					Name: "Warm",
				}
				err = db.Insert(db.Weather).Value(&w)
				if err != nil {
					t.Errorf("Expected a insert weather, got error: %v", err)
				}

				h := Habitat{
					Id:        uuid.New(),
					Name:      "City",
					IdWeather: w.Id,
				}
				err = db.Insert(db.Habitat).Value(&h)
				if err != nil {
					t.Errorf("Expected a insert habitat, got error: %v", err)
				}

				a.IdHabitat = &h.Id
				a.Name = "Update Cat"
				err = db.Update(&db.Animal.IdHabitat, &db.Animal.Name).Where(db.Equals(&db.Animal.Id, a.Id)).Value(a)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				var aselect Animal
				err = db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&aselect)

				if aselect.IdHabitat == nil || *aselect.IdHabitat != h.Id {
					t.Errorf("Expected a update on id habitat, got : %v", aselect.IdHabitat)
				}
				if aselect.Name != "Update Cat" {
					t.Errorf("Expected a update on name, got : %v", aselect.Name)
				}
			},
		},
		{
			desc: "Update_Animal_Pointer",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				w := Weather{
					Name: "Warm",
				}
				err = db.Insert(db.Weather).Value(&w)
				if err != nil {
					t.Errorf("Expected a insert weather, got error: %v", err)
				}

				h := Habitat{
					Id:        uuid.New(),
					Name:      "City",
					IdWeather: w.Id,
				}
				err = db.Insert(db.Habitat).Value(&h)
				if err != nil {
					t.Errorf("Expected a insert habitat, got error: %v", err)
				}

				a.IdHabitat = &h.Id
				a.Name = "Update Cat"
				err = db.Update(&db.Animal.IdHabitat, &db.Animal.Name).Where(db.Equals(&db.Animal.Id, a.Id)).Value(&a)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				var aselect Animal
				err = db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&aselect)

				if aselect.IdHabitat == nil || *aselect.IdHabitat != h.Id {
					t.Errorf("Expected a update on id habitat, got : %v", aselect.IdHabitat)
				}
				if aselect.Name != "Update Cat" {
					t.Errorf("Expected a update on name, got : %v", aselect.Name)
				}
			},
		},
		{
			desc: "Update_Animal_All_Fields",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				w := Weather{
					Name: "Warm",
				}
				err = db.Insert(db.Weather).Value(&w)
				if err != nil {
					t.Errorf("Expected a insert weather, got error: %v", err)
				}

				h := Habitat{
					Id:        uuid.New(),
					Name:      "City",
					IdWeather: w.Id,
				}
				err = db.Insert(db.Habitat).Value(&h)
				if err != nil {
					t.Errorf("Expected a insert habitat, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				a.IdHabitat = &h.Id
				a.Name = "Update Cat"
				err = db.Update(db.Animal).Where(db.Equals(&db.Animal.Id, a.Id)).Value(a)
				if err != nil {
					t.Errorf("Expected a update, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				var aselect Animal
				err = db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&aselect)

				if aselect.IdHabitat == nil || *aselect.IdHabitat != h.Id {
					t.Errorf("Expected a update on id habitat, got : %v", aselect.IdHabitat)
				}
				if aselect.Name != "Update Cat" {
					t.Errorf("Expected a update on name, got : %v", aselect.Name)
				}
			},
		},
		{
			desc: "UpdateIn_AnimalFood",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				f := []Food{{Id: uuid.New(), Name: "Meat"}, {Id: uuid.New(), Name: "Grass"}}
				err = db.Insert(db.Food).Value(&f)
				if err != nil {
					t.Fatalf("Expected insert foods, got error: %v", err)
				}

				animalFoods := []any{
					f[0].Id, a.Id,
				}
				err = db.InsertIn(db.Food, db.Animal).Values(animalFoods)
				if err != nil {
					t.Fatalf("Expected insert animalFoods, got error: %v", err)
				}

				var aselect []struct {
					IdAnimal int
					IdFood   uuid.UUID
					Animal   string
					Food     string
				}
				err = db.Select(&db.Animal.Id, &db.Food.Id, &db.Animal.Name, &db.Food.Name).Join(db.Animal, db.Food).
					Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&aselect)
				if err != nil {
					t.Fatalf("Expected select animal and food, got error: %v", err)
				}

				if len(aselect) != 1 {
					t.Fatalf("Expected select one, got : %v", len(aselect))
				}

				if aselect[0].Animal != a.Name || aselect[0].Food != f[0].Name {
					t.Fatalf("Expected %v and %v, got : %v and %v", a.Name, f[0].Name, aselect[0].Animal, aselect[0].Food)
				}

				err = db.UpdateIn(db.Animal, db.Food).
					Where(db.Equals(&db.Animal.Id, aselect[0].IdAnimal),
						db.And(),
						db.Equals(&db.Food.Id, aselect[0].IdFood)).
					Value(f[1].Id)
				if err != nil {
					t.Fatalf("Expected update animalfood, got error: %v", err)
				}

				err = db.Select(&db.Animal.Id, &db.Food.Id, &db.Animal.Name, &db.Food.Name).Join(db.Animal, db.Food).
					Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&aselect)
				if err != nil {
					t.Fatalf("Expected select animal and food, got error: %v", err)
				}

				if len(aselect) != 1 {
					t.Fatalf("Expected select one, got : %v", len(aselect))
				}

				if aselect[0].Animal != a.Name || aselect[0].Food != f[1].Name {
					t.Fatalf("Expected %v and %v, got : %v and %v", a.Name, f[1].Name, aselect[0].Animal, aselect[0].Food)
				}
			},
		},
		{
			desc: "UpdateIn_AnimalFood_Pointer",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				f := []Food{{Id: uuid.New(), Name: "Meat"}, {Id: uuid.New(), Name: "Grass"}}
				err = db.Insert(db.Food).Value(&f)
				if err != nil {
					t.Fatalf("Expected insert foods, got error: %v", err)
				}

				animalFoods := []any{
					f[0].Id, a.Id,
				}
				err = db.InsertIn(db.Food, db.Animal).Values(animalFoods)
				if err != nil {
					t.Fatalf("Expected insert animalFoods, got error: %v", err)
				}

				var aselect []struct {
					IdAnimal int
					IdFood   uuid.UUID
					Animal   string
					Food     string
				}
				err = db.Select(&db.Animal.Id, &db.Food.Id, &db.Animal.Name, &db.Food.Name).Join(db.Animal, db.Food).
					Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&aselect)
				if err != nil {
					t.Fatalf("Expected select animal and food, got error: %v", err)
				}

				if len(aselect) != 1 {
					t.Fatalf("Expected select one, got : %v", len(aselect))
				}

				if aselect[0].Animal != a.Name || aselect[0].Food != f[0].Name {
					t.Fatalf("Expected %v and %v, got : %v and %v", a.Name, f[0].Name, aselect[0].Animal, aselect[0].Food)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				err = db.UpdateIn(db.Animal, db.Food).
					Where(db.Equals(&db.Animal.Id, aselect[0].IdAnimal),
						db.And(),
						db.Equals(&db.Food.Id, aselect[0].IdFood)).
					Value(&f[1].Id)
				if err != nil {
					t.Fatalf("Expected update animalfood, got error: %v", err)
				}

				if db.ConnPool.Stats().InUse != 0 {
					t.Errorf("Expected closed connection, got: %v", db.ConnPool.Stats().InUse)
				}

				err = db.Select(&db.Animal.Id, &db.Food.Id, &db.Animal.Name, &db.Food.Name).Join(db.Animal, db.Food).
					Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&aselect)
				if err != nil {
					t.Fatalf("Expected select animal and food, got error: %v", err)
				}

				if len(aselect) != 1 {
					t.Fatalf("Expected select one, got : %v", len(aselect))
				}

				if aselect[0].Animal != a.Name || aselect[0].Food != f[1].Name {
					t.Fatalf("Expected %v and %v, got : %v and %v", a.Name, f[1].Name, aselect[0].Animal, aselect[0].Food)
				}
			},
		},
		{
			desc: "Update_Invalid_Tables_1",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = db.Update(&db.Animal.IdHabitat, &db.Food.Name).Where(db.Equals(&db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrTooManyTablesUpdate) {
					t.Errorf("Expected a goe.ErrTooManyTablesUpdate, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Invalid_Tables_2",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = db.Update(db.Animal, db.Food).Where(db.Equals(&db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrTooManyTablesUpdate) {
					t.Errorf("Expected a goe.ErrTooManyTablesUpdate, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Invalid_Arg",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = db.Update(db.DB).Where(db.Equals(&db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected a goe.ErrInvalidArg, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Invalid_Where",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = db.Update(db.Animal).Where(db.Equals(db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrInvalidWhere) {
					t.Errorf("Expected a goe.ErrInvalidWhere, got error: %v", err)
				}
			},
		},
		{
			desc: "Update_Invalid_Value",
			testCase: func(t *testing.T) {
				a := 1
				err = db.Update(db.Animal).Where(db.Equals(&db.Animal.Id, 2)).Value(a)
				if !errors.Is(err, goe.ErrInvalidUpdateValue) {
					t.Errorf("Expected a goe.ErrInvalidUpdateValue, got error: %v", err)
				}
			},
		},
		{
			desc: "UpdateIn_Invalid_Tables",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = db.UpdateIn(db.Animal, db.Flag).Where(db.Equals(db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrNotManyToMany) {
					t.Errorf("Expected a goe.ErrNoMatchesTables, got error: %v", err)
				}
			},
		},
		{
			desc: "UpdateIn_Invalid_Where",
			testCase: func(t *testing.T) {
				a := Animal{
					Name: "Cat",
				}

				a.Name = "Update Cat"
				err = db.UpdateIn(db.Animal, db.Food).Where(db.Equals(db.Animal.Id, a.Id)).Value(a)
				if !errors.Is(err, goe.ErrInvalidWhere) {
					t.Errorf("Expected a goe.ErrInvalidWhere, got error: %v", err)
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
