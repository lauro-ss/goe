package goe

import (
	"context"
	"errors"
	"reflect"
)

var ErrInvalidArg = errors.New("goe: invalid argument. try sending a pointer to a database mapped struct as argument")
var ErrTooManyTablesUpdate = errors.New("goe: invalid table. try sending arguments from the same table")

type Config struct {
	LogQuery bool
}

type DB struct {
	config   *Config
	ConnPool ConnectionPool
	addrMap  map[uintptr]field
	driver   Driver
}

func (db *DB) Migrate(m *Migrator) error {
	c, err := db.ConnPool.Conn(context.Background())
	if err != nil {
		return err
	}
	if m.Error != nil {
		return m.Error
	}
	db.driver.Migrate(m, c)
	return nil
}

// Select creates a select state with passed args
//
// Select uses [context.Background] internally;
// to specify the context, use [DB.SelectContext].
//
// # Example
//
//	// get all fields from animal table
//	// same as "select * from animal;"
//	db.Select(db.Animal).Scan(&a)
//
//	// get animal name and emoji
//	// same as "select name, emoji from animal;"
//	db.Select(&db.Animal.Name, &db.Animal.Emoji).Scan(&a)
//
//	// get a fk id from many to one
//	// same as "select idanimal from status;"
//	db.Select(&db.Status.Animal).Scan(&s)
//
//	// get all foods columns by id animal makeing a join betwent animal and food
//	db.Select(db.Food).Join(db.Animal, db.Food).
//		Where(db.Equals(&db.Animal.Id, "fc1865b4-6f2d-4cc6-b766-49c2634bf5c4")).Scan(&a)
func (db *DB) Select(args ...any) *stateSelect {
	return db.SelectContext(context.Background(), args...)
}

// SelectContext creates a select state with passed args
func (db *DB) SelectContext(ctx context.Context, args ...any) *stateSelect {
	uintArgs, aggregates, err := getArgsSelect(db.addrMap, args...)

	var state *stateSelect
	if err != nil {
		state = createSelectState(nil, db.config, ctx, nil, err)
		return state.querySelect(nil, nil)
	}

	state = createSelectState(db.ConnPool, db.config, ctx, db.driver, err)

	state.addrMap = db.addrMap
	return state.querySelect(uintArgs, aggregates)
}

func (db *DB) Count(arg any) any {
	f := getArg(arg, db.addrMap)
	if f == nil {
		return nil
	}
	return createAggregate("COUNT", f)
}

// Insert creates a insert state for table
//
// Insert uses [context.Background] internally;
// to specify the context, use [DB.InsertContext].
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
func (db *DB) Insert(table any) *stateInsert {
	return db.InsertContext(context.Background(), table)
}

// InsertContext creates a insert state for table
func (db *DB) InsertContext(ctx context.Context, table any) *stateInsert {
	stringArgs, err := getArgs(db.addrMap, table)

	var state *stateInsert
	if err != nil {
		state = createInsertState(nil, db.config, ctx, nil, err)
		return state.queryInsert(nil, nil)
	}

	state = createInsertState(db.ConnPool, db.config, ctx, db.driver, err)

	return state.queryInsert(stringArgs, db.addrMap)
}

// Update creates a update state for table
//
// Update uses [context.Background] internally;
// to specify the context, use [DB.UpdateContext].
//
// # Example
//
//	// update all columns and rows from animal
//	db.Update(db.Animal).Value(a)
//
//	// value can be pointer or value
//	db.Update(db.Animal).Value(&a)
//
//	// update one row and all columns from animal
//	// primary keys auto incremented are ignored
//	db.Update(db.Animal).Where(db.Equals(&db.Animal.Id, a.Id)).Value(a)
//
//	// update one row and column name from animal
//	db.Update(&db.Animal.Name).Where(db.Equals(&db.Animal.Id, a.Id)).Value(a)
func (db *DB) Update(table ...any) *stateUpdate {
	return db.UpdateContext(context.Background(), table...)
}

// UpdateContext creates a update state for table
func (db *DB) UpdateContext(ctx context.Context, table ...any) *stateUpdate {
	stringArgs, err := getArgsUpdate(db.addrMap, table...)

	var state *stateUpdate
	if err != nil {
		state = createUpdateState(nil, db.config, ctx, nil, err)
		return state.queryUpdate(nil, nil)
	}
	state = createUpdateState(db.ConnPool, db.config, ctx, db.driver, err)

	return state.queryUpdate(stringArgs, db.addrMap)
}

// Delete creates a delete state for table
//
// Delete uses [context.Background] internally;
// to specify the context, use [DB.DeleteContext].
//
// # Example
//
//	// delete all rows from status
//	db.Delete(db.Status).Where()
func (db *DB) Delete(table any) *stateDelete {
	return db.DeleteContext(context.Background(), table)
}

// DeleteContext creates a delete state for table
func (db *DB) DeleteContext(ctx context.Context, table any) *stateDelete {
	stringArgs, err := getArgs(db.addrMap, table)

	var state *stateDelete
	if err != nil {
		state = createDeleteState(nil, db.config, ctx, nil, err)
		return state.queryDelete(nil, nil)
	}
	state = createDeleteState(db.ConnPool, db.config, ctx, db.driver, err)

	return state.queryDelete(stringArgs, db.addrMap)
}

func getArg(arg any, addrMap map[uintptr]field) field {
	v := reflect.ValueOf(arg)
	if v.Kind() != reflect.Pointer {
		return nil
	}

	addr := uintptr(v.UnsafePointer())
	if addrMap[addr] != nil {
		return addrMap[addr]
	}
	return nil
}

// Equals creates a "=" to value inside a where clause
//
// # Example
//
//	// delete all rows from AnimalFood that matches the idFood
//	db.DeleteIn(db.Animal, db.Food).Where(db.Equals(&db.Food.Id, "fc1865b4-6f2d-4cc6-b766-49c2634bf5c4"))
func (db *DB) Equals(arg any, value any) operator {
	if a := getArg(arg, db.addrMap); a != nil {
		return a.buildComplexOperator("=", value)
	}
	return nil
}

// NotEquals creates a "<>" to value inside a where clause
//
// # Example
//
//	// get all foods that name are not cookie
//	db.Select(db.Food).Where(db.NotEquals(&db.Food.Name, "Cookie")).Scan(&f)
func (db *DB) NotEquals(arg any, value any) operator {
	if a := getArg(arg, db.addrMap); a != nil {
		return a.buildComplexOperator("<>", value)
	}
	return nil
}

// Greater creates a ">" to value inside a where clause
//
// # Example
//
//	// get all animals that was created after 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).Where(db.Greater(&db.Animal.CreateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func (db *DB) Greater(arg any, value any) operator {
	if a := getArg(arg, db.addrMap); a != nil {
		return a.buildComplexOperator(">", value)
	}
	return nil
}

// GreaterEquals creates a ">=" to value inside a where clause
//
// # Example
//
//	// get all animals that was created in or after 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).Where(db.GreaterEquals(&db.Animal.CreateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func (db *DB) GreaterEquals(arg any, value any) operator {
	if a := getArg(arg, db.addrMap); a != nil {
		return a.buildComplexOperator(">=", value)
	}
	return nil
}

// Less creates a "<" to value inside a where clause
//
// # Example
//
//	// get all animals that was updated before 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).Where(db.Less(&db.Animal.UpdateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func (db *DB) Less(arg any, value any) operator {
	if a := getArg(arg, db.addrMap); a != nil {
		return a.buildComplexOperator("<", value)
	}
	return nil
}

// LessEquals creates a "<=" to value inside a where clause
//
// # Example
//
//	// get all animals that was updated in or before 09 of october 2024 at 11:50AM
//	db.Select(db.Animal).Where(db.LessEquals(&db.Animal.UpdateAt, time.Date(2024, time.October, 9, 11, 50, 00, 00, time.Local))).Scan(&a)
func (db *DB) LessEquals(arg any, value any) operator {
	if a := getArg(arg, db.addrMap); a != nil {
		return a.buildComplexOperator("<=", value)
	}
	return nil
}

// Like creates a "LIKE" to value inside a where clause
//
// # Example
//
//	// get all animals that has a "at" in his name
//	db.Select(db.Animal).Where(db.Like(&db.Animal.Name, "%at%")).Scan(&a)
func (db *DB) Like(arg any, value any) operator {
	if a := getArg(arg, db.addrMap); a != nil {
		return a.buildComplexOperator("LIKE", value)
	}
	return nil
}

// Not creates a "NOT" inside a where clause
//
// # Example
//
//	// get all animals that not has a "at" in his name
//	db.Select(db.Animal).Where(db.Not(db.Like(&db.Animal.Name, "%at%"))).Scan(&a)
func (db *DB) Not(o operator) operator {
	if co, ok := o.(complexOperator); ok {
		co.setNot()
		return co
	}
	return nil
}

// And creates a "AND" inside a where clause
//
// # Example
//
//	// and can connect operations
//	db.UpdateIn(db.Animal, db.Food).Where(
//		db.Equals(&db.Animal.Id, "906f4f1f-49e7-47ee-8954-2d6e0a3354cf"),
//		db.And(),
//		db.Equals(&db.Food.Id, "f023a4e7-34e9-4db2-85e0-efe8d67eea1b")).
//		Value("fc1865b4-6f2d-4cc6-b766-49c2634bf5c4")
func (db *DB) And() operator {
	return simpleOperator{operator: "AND"}
}

// Or creates a "OR" inside a where clause
//
// # Example
//
//	// or can connect operations
//	db.DeleteIn(db.Animal, db.Food).Where(
//		db.Equals(&db.Food.Id, "5ad0e5fc-e9f7-4855-9698-d0c10b996f73"),
//		db.Or(),
//		db.Equals(&db.Animal.Id, "401b5e23-5aa7-435e-ba4d-5c1b2f123596"),
//	)
func (db *DB) Or() operator {
	return simpleOperator{operator: "OR"}
}

func getArgsSelect(addrMap map[uintptr]field, args ...any) ([]uintptr, []aggregate, error) {
	uintArgs := make([]uintptr, 0)
	aggregates := make([]aggregate, 0)
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			if valueOf.Type().Name() != "Time" && valueOf.Kind() == reflect.Struct {
				var fieldOf reflect.Value
				for i := 0; i < valueOf.NumField(); i++ {
					fieldOf = valueOf.Field(i)
					if fieldOf.Kind() == reflect.Slice && fieldOf.Type().Elem().Kind() == reflect.Struct {
						continue
					}
					addr := uintptr(fieldOf.Addr().UnsafePointer())
					if addrMap[addr] != nil {
						uintArgs = append(uintArgs, addr)
					}
				}
			} else {
				uintArgs = append(uintArgs, uintptr(valueOf.Addr().UnsafePointer()))
			}
		} else {
			if a, ok := args[i].(aggregate); ok {
				aggregates = append(aggregates, a)
				continue
			}
			return nil, nil, ErrInvalidArg
		}
	}
	if len(uintArgs) == 0 && len(aggregates) == 0 {
		return nil, nil, ErrInvalidArg
	}
	return uintArgs, aggregates, nil
}

func getArgs(addrMap map[uintptr]field, args ...any) ([]uintptr, error) {
	stringArgs := make([]uintptr, 0)
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			if valueOf.Type().Name() != "Time" && valueOf.Kind() == reflect.Struct {
				var fieldOf reflect.Value
				for i := 0; i < valueOf.NumField(); i++ {
					fieldOf = valueOf.Field(i)
					if fieldOf.Kind() == reflect.Slice && fieldOf.Type().Elem().Kind() == reflect.Struct {
						continue
					}
					addr := uintptr(fieldOf.Addr().UnsafePointer())
					if addrMap[addr] != nil {
						stringArgs = append(stringArgs, addr)
					}
				}
			} else {
				stringArgs = append(stringArgs, uintptr(valueOf.Addr().UnsafePointer()))
			}
		} else {
			return nil, ErrInvalidArg
		}
	}
	if len(stringArgs) == 0 {
		return nil, ErrInvalidArg
	}
	return stringArgs, nil
}

func getArgsUpdate(addrMap map[uintptr]field, args ...any) ([]uintptr, error) {
	stringArgs := make([]uintptr, 0)
	var table string
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			if valueOf.Type().Name() != "Time" && valueOf.Kind() == reflect.Struct {
				var fieldOf reflect.Value
				for i := 0; i < valueOf.NumField(); i++ {
					fieldOf = valueOf.Field(i)
					if fieldOf.Kind() == reflect.Slice && fieldOf.Type().Elem().Kind() == reflect.Struct {
						continue
					}
					addr := uintptr(fieldOf.Addr().UnsafePointer())
					if addrMap[addr] != nil {
						if table != "" && string(addrMap[addr].table()) != table {
							return nil, ErrTooManyTablesUpdate
						}
						table = string(addrMap[addr].table())
						stringArgs = append(stringArgs, addr)
					}
				}
			} else {
				//TODO: Check this, update all comparable table to a Id
				addr := uintptr(valueOf.Addr().UnsafePointer())
				if addrMap[addr] != nil {
					if table != "" && string(addrMap[addr].table()) != table {
						return nil, ErrTooManyTablesUpdate
					}
					table = string(addrMap[addr].table())
					stringArgs = append(stringArgs, uintptr(valueOf.Addr().UnsafePointer()))
				}
			}
		} else {
			return nil, ErrInvalidArg
		}
	}
	if len(stringArgs) == 0 {
		return nil, ErrInvalidArg
	}
	return stringArgs, nil
}

func getArgsIn(addrMap map[uintptr]field, args ...any) ([]uintptr, error) {
	stringArgs := make([]uintptr, 2)
	var ptr uintptr
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			ptr = uintptr(valueOf.Addr().UnsafePointer())
			if addrMap[ptr] != nil {
				stringArgs[i] = ptr
			}
		} else {
			return nil, ErrInvalidArg
		}
	}

	if stringArgs[0] == 0 || stringArgs[1] == 0 {
		return nil, ErrInvalidArg
	}
	return stringArgs, nil
}

func getArgsTables(addrMap map[uintptr]field, args ...any) ([]byte, error) {
	tables := make([]byte, 0)
	var ptr uintptr
	for i := range args {
		if reflect.ValueOf(args[i]).Kind() == reflect.Ptr {
			valueOf := reflect.ValueOf(args[i]).Elem()
			ptr = uintptr(valueOf.Addr().UnsafePointer())
			if addrMap[ptr] == nil {
				//TODO: add ErrInvalidTable
				return nil, ErrInvalidArg
			}
			tables = append(tables, addrMap[ptr].table()...)
		} else {
			return nil, ErrInvalidArg
		}
	}

	return tables, nil
}
