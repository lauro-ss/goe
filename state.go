package goe

import (
	"errors"
	"reflect"
)

var ErrInvalidScan = errors.New("goe: invalid scan target. try sending a pointer for scan")

var ErrInvalidInsertValue = errors.New("goe: invalid insert value. try sending a pointer to a struct as value")
var ErrInvalidInsertBatchValue = errors.New("goe: invalid insert value. try sending a pointer to a slice of struct as value")
var ErrEmptyBatchValue = errors.New("goe: can't insert a empty batch value")
var ErrInvalidInsertPointer = errors.New("goe: invalid insert value. try sending a pointer as value")

var ErrInvalidInsertInValue = errors.New("goe: invalid insertIn value. try sending only two values or a size even slice")

var ErrInvalidUpdateValue = errors.New("goe: invalid update value. try sending a struct or a pointer to struct as value")

type stateSelect struct {
	conn    Connection
	addrMap map[uintptr]field
	builder *builder
	err     error
}

func createSelectState(conn Connection, e error) *stateSelect {
	return &stateSelect{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateSelect) Where(brs ...operator) *stateSelect {
	s.builder.brs = brs
	return s
}

func (s *stateSelect) Join(t1 any, t2 any) *stateSelect {
	args, err := getArgsIn(t1, t2)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.buildSelectJoins(s.addrMap, "JOIN", args)
	return s
}

func (s *stateSelect) InnerJoin(t1 any, t2 any) *stateSelect {
	args, err := getArgsIn(t1, t2)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.buildSelectJoins(s.addrMap, "INNER JOIN", args)
	return s
}

func (s *stateSelect) RightJoin(t1 any, t2 any) *stateSelect {
	args, err := getArgsIn(t1, t2)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.buildSelectJoins(s.addrMap, "RIGHT JOIN", args)
	return s
}

func (s *stateSelect) LeftJoin(t1 any, t2 any) *stateSelect {
	args, err := getArgsIn(t1, t2)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.buildSelectJoins(s.addrMap, "LEFT JOIN", args)
	return s
}

func (s *stateSelect) querySelect(args []uintptr) *stateSelect {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildSelect(s.addrMap)
	}
	return s
}

// Scan fills the target with the returned sql data,
// target can be a pointer or a pointer to [Slice].
//
// In case of passing a pointer of struct or a pointer to slice of
// struct, goe package will match the fields by name
//
// Scan uses [sql.Row] if a not slice pointer is the target, in
// this case can return [sql.ErrNoRows]
//
// Scan returns the SQL generated and a nil error if succeed.
//
// # Example:
//
//	// using struct
//	var a Animal
//	db.Select(db.Animal).Scan(&a)
//
//	// using slice
//	var a []Animal
//	db.Select(db.Animal).Scan(&a)
func (s *stateSelect) Scan(target any) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		return "", ErrInvalidScan
	}

	//generate query
	s.err = s.builder.buildSqlSelect()
	if s.err != nil {
		return "", s.err
	}

	sql := s.builder.sql.String()
	return sql, handlerResult(s.conn, sql, value.Elem(), s.builder.argsAny, s.builder.structColumns)
}

/*
State Insert
*/
type stateInsert struct {
	conn    Connection
	builder *builder
	err     error
}

func createInsertState(conn Connection, e error) *stateInsert {
	return &stateInsert{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateInsert) queryInsert(args []uintptr, addrMap map[uintptr]field) *stateInsert {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildInsert(addrMap)
	}
	return s
}

// Value inserts the value inside the database, and updates the Id field if
// is a auto increment.
//
// The value needs to be a pointer to a struct of database types
// or a pointer to slice of database types (in case of batch insert).
//
// Value returns the SQL generated and error as nil if insert with success.
//
// # Example
//
//	// insert one value
//	food := Food{Id: "fc1865b4-6f2d-4cc6-b766-49c2634bf5c4", Name: "Cookie", Emoji: "🍪"}
//	db.Insert(db.Food).Value(&food)
//
//	// insert batch values
//	foods := []Food{
//		{Id: "401b5e23-5aa7-435e-ba4d-5c1b2f123596", Name: "Meat", Emoji: "🥩"},
//		{Id: "f023a4e7-34e9-4db2-85e0-efe8d67eea1b", Name: "Hotdog", Emoji: "🌭"},
//		{Id: "fc1865b4-6f2d-4cc6-b766-49c2634bf5c4", Name: "Cookie", Emoji: "🍪"},
//	}
//	db.Insert(db.Food).Value(&foods)
func (s *stateInsert) Value(value any) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	v := reflect.ValueOf(value)

	if v.Kind() != reflect.Ptr {
		return "", ErrInvalidInsertPointer
	}

	v = v.Elem()

	if v.Kind() == reflect.Slice {
		return s.batchValue(v)
	}

	if v.Kind() != reflect.Struct {
		return "", ErrInvalidInsertValue
	}

	idName := s.builder.buildValues(v)

	sql := s.builder.sql.String()
	return sql, handlerValuesReturning(s.conn, sql, v, s.builder.argsAny, idName)
}

func (s *stateInsert) batchValue(value reflect.Value) (string, error) {
	if value.Len() == 0 {
		return "", ErrEmptyBatchValue
	}

	if value.Index(0).Kind() != reflect.Struct {
		return "", ErrInvalidInsertBatchValue
	}
	idName := s.builder.buildBatchValues(value)

	sql := s.builder.sql.String()
	return sql, handlerValuesReturningBatch(s.conn, sql, value, s.builder.argsAny, idName)
}

type stateInsertIn struct {
	conn    Connection
	builder *builder
	err     error
}

func createInsertStateIn(conn Connection, e error) *stateInsertIn {
	return &stateInsertIn{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateInsertIn) queryInsertIn(args []uintptr, addrMap map[uintptr]field) *stateInsertIn {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildInsertIn(addrMap)
	}
	return s
}

// Values inserts the values inside the database.
//
// The values needs to be the same type as the ids from the database tables.
// Values can be a even slice of [any] with the positional matchs of the ids
//
// Values returns the SQL generated and error as nil if insert with success.
//
// # Example:
//
//	// insert into AnimalFood first value is for idFood and second is for idAnimal
//	db.InsertIn(db.Food, db.Animal).Values("5ad0e5fc-e9f7-4855-9698-d0c10b996f73", "401b5e23-5aa7-435e-ba4d-5c1b2f123596")
//
//	// insert into AnimalHabitat, first value is a uuid for idAnimal and second is a int for idHabitat
//	db.InsertIn(db.Animal, db.Habitat).Values("5ad0e5fc-e9f7-4855-9698-d0c10b996f73", 25)
func (s *stateInsertIn) Values(v ...any) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	switch len(v) {
	case 1:
		value := reflect.ValueOf(v[0])
		if value.Kind() == reflect.Pointer {
			value = value.Elem()
		}
		if value.Kind() != reflect.Slice || value.Len() < 2 || value.Len()%2 != 0 {
			return "", ErrInvalidInsertInValue
		}

		s.err = s.builder.buildValuesInBatch(value)
		if s.err != nil {
			return "", s.err
		}

		sql := s.builder.sql.String()
		return sql, handlerValues(s.conn, sql, s.builder.argsAny)
	case 2:
		s.builder.argsAny = append(s.builder.argsAny, v[0])
		s.builder.argsAny = append(s.builder.argsAny, v[1])

		s.err = s.builder.buildValuesIn()
		if s.err != nil {
			return "", s.err
		}

		sql := s.builder.sql.String()
		return sql, handlerValues(s.conn, sql, s.builder.argsAny)
	default:
		return "", ErrInvalidInsertInValue
	}
}

/*
State Update
*/
type stateUpdate struct {
	conn    Connection
	builder *builder
	err     error
}

func createUpdateState(conn Connection, e error) *stateUpdate {
	return &stateUpdate{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateUpdate) Where(brs ...operator) *stateUpdate {
	s.builder.brs = brs
	return s
}

func (s *stateUpdate) queryUpdate(args []uintptr, addrMap map[uintptr]field) *stateUpdate {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildUpdate(addrMap)
	}
	return s
}

// Value updates the targets in the database.
//
// The value can be a pointer to struct or a struct value.
//
// Value returns the SQL generated and error as nil if update with success.
//
// # Example
//
//	// updates all rows with aStruct values
//	db.Update(db.Animal).Value(aStruct)
//
//	// updates single row using where
//	db.Update(db.Animal).Where(db.Equals(&db.Animal.Id, aStruct.Id)).Value(aStruct)
func (s *stateUpdate) Value(value any) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", ErrInvalidUpdateValue
	}

	s.builder.buildSet(v)

	//generate query
	s.err = s.builder.buildSqlUpdate()
	if s.err != nil {
		return "", s.err
	}

	sql := s.builder.sql.String()
	return sql, handlerValues(s.conn, sql, s.builder.argsAny)
}

type stateUpdateIn struct {
	conn    Connection
	builder *builder
	err     error
}

func createUpdateInState(conn Connection, e error) *stateUpdateIn {
	return &stateUpdateIn{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateUpdateIn) Where(brs ...operator) *stateUpdateIn {
	s.builder.brs = brs
	return s
}

func (s *stateUpdateIn) queryUpdateIn(args []uintptr, addrMap map[uintptr]field) *stateUpdateIn {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildUpdateIn(addrMap)
	}
	return s
}

// Value updates the targets in the database.
//
// The value needs to be the same as the id from database.
//
// Value returns the SQL generated and error as nil if update with success.
//
// # Example
//
//	// updates idFood in the matched where
//	db.UpdateIn(db.Animal, db.Food).Where(
//		db.Equals(&db.Animal.Id, "906f4f1f-49e7-47ee-8954-2d6e0a3354cf"),
//		db.And(),
//		db.Equals(&db.Food.Id, "f023a4e7-34e9-4db2-85e0-efe8d67eea1b")).
//		Value("fc1865b4-6f2d-4cc6-b766-49c2634bf5c4")
func (s *stateUpdateIn) Value(value any) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	s.builder.argsAny = append(s.builder.argsAny, value)

	s.err = s.builder.buildSetIn()
	if s.err != nil {
		return "", s.err
	}

	s.err = s.builder.buildSqlUpdateIn()
	if s.err != nil {
		return "", s.err
	}

	sql := s.builder.sql.String()
	return sql, handlerValues(s.conn, sql, s.builder.argsAny)
}

type stateDelete struct {
	conn    Connection
	builder *builder
	err     error
}

func createDeleteState(conn Connection, e error) *stateDelete {
	return &stateDelete{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateDelete) queryDelete(args []uintptr, addrMap map[uintptr]field) *stateDelete {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildDelete(addrMap)
	}
	return s
}

// Where from state delete executes the delete command in the database.
//
// Returns the SQL generated and error as nil if delete with success.
//
// # Example
//
//	// delete all animals
//	db.Delete(db.Animal).Where()
//
//	// delete matched animals
//	db.Delete(db.Animal).Where(db.Equals(&db.Animal.Id, "906f4f1f-49e7-47ee-8954-2d6e0a3354cf"))
func (s *stateDelete) Where(brs ...operator) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	s.builder.brs = brs

	s.err = s.builder.buildSqlDelete()
	if s.err != nil {
		return "", s.err
	}

	sql := s.builder.sql.String()
	return sql, handlerValues(s.conn, sql, s.builder.argsAny)
}

type stateDeleteIn struct {
	conn    Connection
	builder *builder
	err     error
}

func createDeleteInState(conn Connection, e error) *stateDeleteIn {
	return &stateDeleteIn{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateDeleteIn) queryDeleteIn(args []uintptr, addrMap map[uintptr]field) *stateDeleteIn {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildDeleteIn(addrMap)
	}
	return s
}

// Where from state delete executes the delete command in the database.
//
// Returns the SQL generated and error as nil if delete with success.
//
// # Example
//
//	// delete all rows from AnimalFood
//	db.DeleteIn(db.Animal, db.Food).Where()
//
//	// delete all matched rows from AnimalFood
//	db.DeleteIn(db.Animal, db.Food).Where(
//		db.Equals(&db.Food.Id, "5ad0e5fc-e9f7-4855-9698-d0c10b996f73"),
//		db.Or(),
//		db.Equals(&db.Animal.Id, "401b5e23-5aa7-435e-ba4d-5c1b2f123596"),
//	)
func (s *stateDeleteIn) Where(brs ...operator) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	s.builder.brs = brs

	s.err = s.builder.buildSqlDeleteIn()
	if s.err != nil {
		return "", s.err
	}

	sql := s.builder.sql.String()
	return sql, handlerValues(s.conn, sql, s.builder.argsAny)
}
