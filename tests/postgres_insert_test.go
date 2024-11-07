package tests_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
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
				f := Food{Name: "Meat"}
				err = db.Insert(db.Food).Value(&f)
				if err != nil {
					t.Errorf("Expected a insert food, got error: %v", err)
				}
				if f.Id == 0 {
					t.Errorf("Expected a Id value, got : %v", a.Id)
				}

				err = db.InsertIn(db.Animal, db.Food).Values(a.Id, f.Id)
				if err != nil {
					t.Errorf("Expected a insert AnimalFood, got error: %v", err)
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
