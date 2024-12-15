package wh

type Operation struct {
	Arg      any
	Value    any
	Operator string
}

// Equals creates a "=" to value inside a where clause
//
// # Example
//
//	// delete Food with Id fc1865b4-6f2d-4cc6-b766-49c2634bf5c4
//	db.Delete(db.Food).Where(wh.Equals(&db.Food.Id, "fc1865b4-6f2d-4cc6-b766-49c2634bf5c4"))
//
//	// implicit join using where equals
//	db.Select(db.Animal).
//	From(db.Animal, db.AnimalFood, db.Food).
//	Where(
//		wh.Equals(&db.Animal.Id, &db.AnimalFood.IdAnimal),
//		wh.And(),
//		wh.Equals(&db.Food.Id, &db.AnimalFood.IdFood)).
//	Scan(&a)
func Equals(a any, v any) Operation {
	return Operation{Arg: a, Value: v, Operator: "="}
}

// NotEquals creates a "<>" to value inside a where clause
//
// # Example
//
//	// get all foods that name are not Cookie
//	db.Select(db.Food).From(db.Animal).
//	Where(wh.NotEquals(&db.Food.Name, "Cookie")).Scan(&f)
func NotEquals(a any, v any) Operation {
	return Operation{Arg: a, Value: v, Operator: "<>"}
}

// Greater creates a ">" to value inside a where clause
//
// # Example
//
//	// get all animals that was created after 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.Greater(&db.Animal.CreateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func Greater(a any, v any) Operation {
	return Operation{Arg: a, Value: v, Operator: ">"}
}

// GreaterEquals creates a ">=" to value inside a where clause
//
// # Example
//
//	// get all animals that was created in or after 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.GreaterEquals(&db.Animal.CreateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func GreaterEquals(a any, v any) Operation {
	return Operation{Arg: a, Value: v, Operator: ">="}
}

// Less creates a "<" to value inside a where clause
//
// # Example
//
//	// get all animals that was updated before 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.Less(&db.Animal.UpdateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func Less(a any, v any) Operation {
	return Operation{Arg: a, Value: v, Operator: "<"}
}

// LessEquals creates a "<=" to value inside a where clause
//
// # Example
//
//	// get all animals that was updated in or before 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).From(db.Animal).
//	Where(wh.LessEquals(&db.Animal.UpdateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func LessEquals(a any, v any) Operation {
	return Operation{Arg: a, Value: v, Operator: "<="}
}

// Like creates a "LIKE" to value inside a where clause
//
// # Example
//
//	// get all animals that has a "at" in his name
//	db.Select(db.Animal).From(db.Animal).Where(wh.Like(&db.Animal.Name, "%at%")).Scan(&a)
func Like(a any, v any) Operation {
	return Operation{Arg: a, Value: v, Operator: "LIKE"}
}

// Not creates a "NOT" inside a where clause
//
// # Example
//
//	// get all animals that not has a "at" in his name
//	db.Select(db.Animal).From(db.Animal).Where(wh.Not(wh.Like(&db.Animal.Name, "%at%"))).Scan(&a)
func Not(o Operation) Operation {
	o.Operator = "NOT " + o.Operator
	return o
}

type Logical struct {
	Operator string
}

// And creates a "AND" inside a where clause
//
// # Example
//
//	// and can connect operations
//	db.Update(db.Animal).Where(
//		wh.Equals(&db.Animal.Status, "Eating"),
//		wh.And(),
//		wh.Like(&db.Animal.Name, "%Cat%")).
//		Value(a)
func And() Logical {
	return Logical{Operator: "AND"}
}

// Or creates a "OR" inside a where clause
//
// # Example
//
//	// or can connect operations
//	db.Update(db.Animal).Where(
//		wh.Equals(&db.Animal.Status, "Eating"),
//		wh.Or(),
//		wh.Like(&db.Animal.Name, "%Cat%")).
//		Value(a)
func Or() Logical {
	return Logical{Operator: "OR"}
}
