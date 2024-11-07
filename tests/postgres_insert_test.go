package tests_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lauro-ss/goe"
)

func TestPostgresInsert(t *testing.T) {
	db, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Insert_Flag",
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

				var fs Flag
				db.Select(db.Flag).Where(db.Equals(&db.Flag.Id, f.Id)).Scan(&fs)

				if fs.Id != f.Id {
					t.Errorf("Expected %v, got : %v", f.Id, fs.Id)
				}

				if fs.Name != f.Name {
					t.Errorf("Expected %v, got : %v", f.Name, fs.Name)
				}

				if fs.Float32 != f.Float32 {
					t.Errorf("Expected %v, got : %v", f.Float32, fs.Float32)
				}
				if fs.Float64 != f.Float64 {
					t.Errorf("Expected %v, got : %v", f.Float64, fs.Float64)
				}

				if fs.Today.Second() != f.Today.Second() {
					t.Errorf("Expected %v, got : %v", f.Today, fs.Today)
				}

				if fs.Int != f.Int {
					t.Errorf("Expected %v, got : %v", f.Int, fs.Int)
				}
				if fs.Int8 != f.Int8 {
					t.Errorf("Expected %v, got : %v", f.Int8, fs.Int8)
				}
				if fs.Int16 != f.Int16 {
					t.Errorf("Expected %v, got : %v", f.Int16, fs.Int16)
				}
				if fs.Int32 != f.Int32 {
					t.Errorf("Expected %v, got : %v", f.Int32, fs.Int32)
				}
				if fs.Int64 != f.Int64 {
					t.Errorf("Expected %v, got : %v", f.Int64, fs.Int64)
				}

				if fs.Uint != f.Uint {
					t.Errorf("Expected %v, got : %v", f.Uint, fs.Uint)
				}
				if fs.Uint8 != f.Uint8 {
					t.Errorf("Expected %v, got : %v", f.Uint8, fs.Uint8)
				}
				if fs.Uint16 != f.Uint16 {
					t.Errorf("Expected %v, got : %v", f.Uint16, fs.Uint16)
				}
				if fs.Uint32 != f.Uint32 {
					t.Errorf("Expected %v, got : %v", f.Uint32, fs.Uint32)
				}
				if fs.Uint64 != f.Uint64 {
					t.Errorf("Expected %v, got : %v", f.Uint64, fs.Uint64)
				}

				if fs.Bool != f.Bool {
					t.Errorf("Expected %v, got : %v", f.Bool, fs.Bool)
				}
			},
		},
		{
			desc: "Insert_Animal",
			testCase: func(t *testing.T) {
				a := Animal{Name: "Cat"}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert, got error: %v", err)
				}
				if a.Id == 0 {
					t.Errorf("Expected a Id value, got : %v", a.Id)
				}
			},
		},
		{
			desc: "InsertIn_AnimalFood",
			testCase: func(t *testing.T) {
				a := Animal{Name: "Cat"}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}
				if a.Id == 0 {
					t.Errorf("Expected a Id value, got : %v", a.Id)
				}
				f := Food{Id: uuid.New(), Name: "Meat"}
				err = db.Insert(db.Food).Value(&f)
				if err != nil {
					t.Errorf("Expected a insert food, got error: %v", err)
				}

				err = db.InsertIn(db.Animal, db.Food).Values(a.Id, f.Id)
				if err != nil {
					t.Errorf("Expected a insert AnimalFood, got error: %v", err)
				}
			},
		},
		{
			desc: "Insert_Batch_Animal",
			testCase: func(t *testing.T) {
				animals := []Animal{
					{Name: "Cat"},
					{Name: "Dog"},
					{Name: "Forest Cat"},
					{Name: "Bear"},
					{Name: "Lion"},
					{Name: "Puma"},
					{Name: "Snake"},
					{Name: "Whale"},
				}
				err = db.Insert(db.Animal).Value(&animals)
				if err != nil {
					t.Fatalf("Expected insert animals, got error: %v", err)
				}
				for i := range animals {
					if animals[i].Id == 0 {
						t.Errorf("Expected a Id value, got : %v", animals[i].Id)
					}
				}
			},
		},
		{
			desc: "InsertIn_Batch_Animal",
			testCase: func(t *testing.T) {
				animals := []Animal{
					{Name: "Cat"},
					{Name: "Dog"},
				}
				err = db.Insert(db.Animal).Value(&animals)
				if err != nil {
					t.Fatalf("Expected insert animals, got error: %v", err)
				}

				foods := []Food{
					{Id: uuid.New(), Name: "Meat"},
					{Id: uuid.New(), Name: "Grass"},
				}
				err = db.Insert(db.Food).Value(&foods)
				if err != nil {
					t.Fatalf("Expected insert foods, got error: %v", err)
				}

				animalFoods := []any{
					foods[0].Id, animals[0].Id,
					foods[0].Id, animals[1].Id}
				err = db.InsertIn(db.Food, db.Animal).Values(animalFoods)
				if err != nil {
					t.Fatalf("Expected insert animalFoods, got error: %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Pointer",
			testCase: func(t *testing.T) {
				a := Animal{Name: "Cat"}
				err = db.Insert(db.Animal).Value(a)
				if !errors.Is(err, goe.ErrInvalidInsertPointer) {
					t.Errorf("Expected goe.ErrInvalidInsertPointer, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Value",
			testCase: func(t *testing.T) {
				a := 2
				err = db.Insert(db.Animal).Value(&a)
				if !errors.Is(err, goe.ErrInvalidInsertValue) {
					t.Errorf("Expected goe.ErrInvalidInsertValue, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Batch_Value",
			testCase: func(t *testing.T) {
				animals := []int{
					1,
					2,
				}
				err = db.Insert(db.Animal).Value(&animals)
				if !errors.Is(err, goe.ErrInvalidInsertBatchValue) {
					t.Errorf("Expected goe.ErrInvalidInsertBatchValue, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Empty_Batch",
			testCase: func(t *testing.T) {
				animals := []Animal{}
				err = db.Insert(db.Animal).Value(&animals)
				if !errors.Is(err, goe.ErrEmptyBatchValue) {
					t.Errorf("Expected goe.ErrInvalidInsertBatchValue, got : %v", err)
				}
			},
		},
		{
			desc: "Insert_Invalid_Arg",
			testCase: func(t *testing.T) {
				a := 2
				err = db.Insert(db.DB).Value(&a)
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected goe.ErrInvalidArg, got : %v", err)
				}

				err = db.Insert(nil).Value(&a)
				if !errors.Is(err, goe.ErrInvalidArg) {
					t.Errorf("Expected goe.ErrInvalidArg, got : %v", err)
				}
			},
		},
		{
			desc: "InsertIn_Invalid_Tables",
			testCase: func(t *testing.T) {
				af := []any{
					1, 2,
					3, 4,
				}
				err = db.InsertIn(db.Animal, db.Flag).Values(&af)
				if !errors.Is(err, goe.ErrNoMatchesTables) {
					t.Errorf("Expected goe.ErrNoMatchesTables, got : %v", err)
				}
			},
		},
		{
			desc: "InsertIn_Invalid_Tables_Many_To_Many",
			testCase: func(t *testing.T) {
				af := []any{
					1, 2,
					3, 4,
				}
				err = db.InsertIn(db.Animal, db.Habitat).Values(&af)
				if !errors.Is(err, goe.ErrNotManyToMany) {
					t.Errorf("Expected goe.ErrNotManyToMany, got : %v", err)
				}
			},
		},
		{
			desc: "InsertIn_Invalid_Value",
			testCase: func(t *testing.T) {
				af := []any{
					1, 2,
					3,
				}
				err = db.InsertIn(db.Animal, db.Food).Values(af)
				if !errors.Is(err, goe.ErrInvalidInsertInValue) {
					t.Errorf("Expected goe.ErrInvalidInsertInValue, got : %v", err)
				}

				a := 1
				err = db.InsertIn(db.Animal, db.Food).Values(a)
				if !errors.Is(err, goe.ErrInvalidInsertInValue) {
					t.Errorf("Expected goe.ErrInvalidInsertInValue, got : %v", err)
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
