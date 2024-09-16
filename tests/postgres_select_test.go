package tests_test

import "testing"

func TestPostgresSelect(t *testing.T) {
	db, _ := SetupPostgres()

	err := db.Delete(db.Animal).Where()
	if err != nil {
		t.Fatalf("Expected delete animals, got error: %v", err)
	}
	err = db.Delete(db.Food).Where()
	if err != nil {
		t.Fatalf("Expected delete foods, got error: %v", err)
	}

	animals := []Animal{{Name: "Cat"}, {Name: "Dog"}, {Name: "Forest Cat"}}
	err = db.Insert(db.Animal).Value(&animals)
	if err != nil {
		t.Fatalf("Expected insert animals, got error: %v", err)
	}

	foods := []Food{{Name: "Meat"}, {Name: "Grass"}}
	err = db.Insert(db.Food).Value(&foods)
	if err != nil {
		t.Fatalf("Expected insert foods, got error: %v", err)
	}

	db.InsertIn(db.Animal, db.Food).Values([]int{foods[0].Id, animals[0].Id})

	testCases := []struct {
		desc     string
		testCase func(t *testing.T)
	}{
		{
			desc: "Select_animals",
			testCase: func(t *testing.T) {
				var a []Animal
				db.Select(db.Animal).Scan(&a)
				if len(a) != len(animals) {
					t.Errorf("Expected %v animals, got %v", len(animals), len(a))
				}
			},
		},
		{
			desc: "Select_Where_Equals",
			testCase: func(t *testing.T) {
				var a Animal
				db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, animals[0].Id)).Scan(&a)
				if a.Name != animals[0].Name {
					t.Errorf("Expected a %v, got %v", animals[0].Name, a.Name)
				}
			},
		},
		{
			desc: "Select_Where_Like",
			testCase: func(t *testing.T) {
				var a []Animal
				db.Select(db.Animal).Where(db.Like(&db.Animal.Name, "%Cat%")).Scan(&a)
				if len(a) != 2 {
					t.Errorf("Expected %v animals, got %v", 2, len(a))
				}
			},
		},
		{
			desc: "Select_Join",
			testCase: func(t *testing.T) {
				var a []Animal
				db.Select(db.Animal).Join(db.Animal, db.Food).Scan(&a)
				if len(a) != 1 {
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
				db.Select(db.Food).Join(db.Animal, db.Food).Where(db.Equals(&db.Animal.Name, animals[0].Name)).Scan(&f)
				if len(f) != 1 {
					t.Errorf("Expected 1 food, got %v", len(f))
				}
				if f[0].Name != foods[0].Name {
					t.Errorf("Expected %v, got %v", foods[0].Name, f[0].Name)
				}
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
