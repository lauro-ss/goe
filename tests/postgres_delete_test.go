package tests_test

import (
	"testing"
)

func TestPostgresDelete(t *testing.T) {
	db, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expected database, got error: %v", err)
	}

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Delete_One_Record",
			testCase: func(t *testing.T) {
				a := Animal{Name: "Dog"}
				err = db.Insert(db.Animal).Value(&a)
				if err != nil {
					t.Errorf("Expected a insert animal, got error: %v", err)
				}

				var as Animal
				err = db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&as)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}

				err = db.Delete(db.Animal).Where(db.Equals(&db.Animal.Id, as.Id))
				if err != nil {
					t.Errorf("Expected a delete animal, got error: %v", err)
				}

				err = db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&as)
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}

				var id int
				err = db.Select(&db.Animal.Id).Where(db.Equals(&db.Animal.Id, a.Id)).Scan(&id)
				if err != nil {
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
				if err != nil {
					t.Errorf("Expected a select, got error: %v", err)
				}

				if len(animals) != 0 {
					t.Errorf(`Expected to delete all "Cat" animals, got: %v`, len(animals))
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
