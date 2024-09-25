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

	foods := []Food{{Name: "Meat"}, {Name: "Grass"}}
	err = db.Insert(db.Food).Value(&foods)
	if err != nil {
		t.Fatalf("Expected insert foods, got error: %v", err)
	}

	animalFoods := []int{
		foods[0].Id, animals[0].Id,
		foods[0].Id, animals[1].Id}
	db.InsertIn(db.Food, db.Animal).Values(animalFoods)

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
	}
	for _, tC := range testCases {
		t.Run(tC.desc, tC.testCase)
	}
}
