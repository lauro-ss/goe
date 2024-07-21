# GOE
 A SQL Like code first ORM for Go (golang)


[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/lauro-ss/goe)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

![](goe.png)
*GOE logo by [Luanexs](https://www.instagram.com/luanexs/)*

## Content
- [Install](#install)
- [Available Drivers](#available-drivers)
    - [PostgreSQL](#postgresql)
- [Quick Start](#quick-start)
- [Database](#database)
	- [Struct Mapping](#struct-mapping)
	- [Setting primary key](#setting-primary-key)
	- [Setting type](#setting-type)
	- [Relationship](#relationship)
		- [One to One](#one-to-one)
		- [Many to One](#many-to-one)
		- [Many to Many](#many-to-many)
	- [Index](#index)
		- [Create Index](#create-index)
		- [Unique Index](#unique-index)
		- [Function Index](#function-index)
		- [Two Columns Index](#two-columns-index)
- [Select](#select)
	- [Select From](#select-from)
	- [Select Specific Fields](#select-specific-fields)
	- [Select Where](#select-where)
	- [Select Join](#select-join)
	- [Pagination](#pagination)
- [Insert](#insert)
- [Update](#update)
- [Delete](#delete)

## Install
```
go get github.com/lauro-ss/goe
```
> As any database/sql support in go, you have to get a specific driver for your database, check [Available Drivers](#available-drivers)

## Available Drivers
### PostgreSQL
```
go get github.com/lauro-ss/postgres
```
## Quick Start
```
package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lauro-ss/goe"
	"github.com/lauro-ss/postgres"
)

// By default field "Id" is primary key
// and all integers types are auto increment
type Animal struct {
	Id    uint
	Name  string
	Emoji string
}

// In goe, it's necessary to define a Database struct
// that implements *goe.DB and all
// the structs that's it's to be mappend
//
// It's through the Database struct that you will
// interact with your database
type Database struct {
	Animal *Animal
	*goe.DB
}

func main() {
	db := &Database{DB: &goe.DB{}}

	DNS := "user=app password=123456 host=localhost port=5432 database=appanimal"
	err := goe.Open(db, postgres.Open(DNS))
	if err != nil {
		fmt.Println("A error ocurred when opening the database", err)
		return
	}

	// migrate all the database tables and print SQL
	db.Migrate(goe.MigrateFrom(db))

	var a Animal
	_, err = db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, 1)).Scan(&a)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fmt.Println("A error occurred when select animal", err)
		return
	}
	if a.Id == 0 {
		a.Name = "Elephant"
		a.Emoji = "🐘"
		_, err := db.Insert(db.Animal).Value(&a)
		if err != nil {
			fmt.Println("A error occurred when inserting animal", err)
			return
		}
	}
	var animals []Animal
	sql, _ := db.Select(db.Animal).Scan(&animals)
	fmt.Println(sql)
	fmt.Println(animals)
}
```
## Database
```
type Database struct {
	User    *User
	Role    *Role
	UserLog *UserLog
	*goe.DB
}
```
> In goe, it's necessary to define a Database struct,
this struct implements *goe.DB and a pointer to all
the structs that's it's to be mappend

> It's through the Database struct that you will
interact with your database
### Struct mapping
```
type User struct {
	Id       uint
	Name     string
	Password string
}
```
> By default the field "Id" is primary key and all ids of integers are auto increment
### Setting primary key
```
type User struct {
	Id       uint `goe:"pk"`
	Name     string
	Password string
}
```
> In case you want to specify 
a primary key use the tag value "pk"
### Setting type
```
type User struct {
	Id       string `goe:"pk;type:uuid"`
	Name     string `goe:"type:varchar(50)"`
	Password string
}
```
> You can specify a type using the tag value "type"
### Relationship
#### One To One
```
type User struct {
	Id       uint
	Name     string
	Password string
}

type UserLog struct {
	Id       uint
	Action   string
	DateTime time.Time
	IdUser   uint `goe:"table:User"`
}
```
> User has one UserLog

> In goe it's necessary to use the tag value "table" in a field named with the pattern "Id + Table"
#### Many To One
```
type User struct {
	Id       uint
	Name     string
	Password string
	UserLogs []UserLog
}

type UserLog struct {
	Id       uint
	Action   string
	DateTime time.Time
	IdUser   uint `goe:"table:User"`
}
```
> User has many UserLog

> The difference from one to one and many to one it's the add of a slice field on the "many" struct

> In goe it's necessary to use the tag value "table" in a field named with the pattern "Id + Table"
#### Many to Many
```
type User struct {
	Id       uint
	Name     string
	Password string
	Roles    []Role `goe:"table:UserRole"`
}

type Role struct {
	Id    uint
	Name  string
	Users []User `goe:"table:UserRole"`
}
```
> One user has many roles and one role has many users

> In goe it's necessary to use the tag value "table" in both slice fields, over the hood goe will create a table "UserRole"

> To manipulete the table "UserRole" use the methods with signature "In", like "InsertIn", "UpdateIn" and "DeleteIn"
### Index
#### Create Index
```
type User struct {
	Id       uint
	Name     string
	Email 	 string `goe:"index(n:idx_email)"`
}
```
> To create a index you need to use the function tag index(), "n" is the parameter for name
#### Unique Index
```
type User struct {
	Id       uint
	Name     string
	Email    string `goe:"index(unique n:idx_email)"`
}
```
> To create a unique index you need to pass the "unique" word inside index()
#### Function Index
```
type User struct {
	Id       uint
	Name     string
	Email    string `goe:"index(n:idx_email f:lower)"`
}
```
> To create a function index you need to pass the "f" parameter with the function name
#### Two Columns Index
```
type User struct {
	Id       uint
	Name    string `goe:"index(unique n:idx_name_email f:lower)"`
	Email   string `goe:"index(unique n:idx_name_email f:lower)"`
}
```
> To create a two columns index it's necessary to inform the index name in both columns

> If you want to create a index for email and mantain the two columns index, just write a comma and you can create a index in the way you want.

```
`goe:"index(unique n:idx_name_email f:lower, n:idx_email)"`
```
## Select
### Select From
```
db.Select(db.Animal).Scan(&a)
```
### Select Specific Fields
```
db.Select(&db.Animal.Name, &db.Animal.Emoji).Scan(&a)
```
### Select Where
```
db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, 1)).Scan(&a)
```
### Select Join
```
db.Select(db.Food).Join(db.Animal, db.Food).
    Where(db.Equals(&db.Animal.Id, 1)).Scan(&a)
```
> Join will throw an error if the tables don't have a many to many or many to one relationship.

> Using join you can access Animal Id for getting foods by Animal.
### Pagination
```
db.Select(db.Animal).Page(2, 20).OrderByAsc(&db.Animal.Name).Scan(&animals)
```
> Get 20 habitats from second page ordered by name

same result with
```
db.Select(db.Animal).Skip(20).Take(20).OrderByAsc(&db.Animal.Name).Scan(&animals)
```
## Insert
### Insert Table
```
var a Animal
a.Name = "Elephant"
a.Emoji = "🐘"
db.Insert(db.Animal).Value(&a)
```
### Insert Batch Table
```
foods := []Food{
		{Name: "Meat", Emoji: "🥩"},
		{Name: "Hotdog", Emoji: "🌭"},
		{Name: "Cookie", Emoji: "🍪"},
	}
db.Insert(db.Food).Value(&foods)
```
### Insert With Foreign Key
```
var a Animal
a.Name = "Elephant"
a.Emoji = "🐘"
a.IdStatus = 1
db.Insert(db.Animal).Value(&a)
```
### Insert Many To Many Table
```
db.InsertIn(db.Food, db.Animal).Values(10, 20)
```
> InsertIn is used for insert values inside many to many tables

> "10" is the value for IdFood and "20" is the value for IdAnimal
### Insert Batch Many To Many Table
```
ids := []uint{
		1, 1,
		1, 2,
		1, 3,
	}
db.InsertIn(db.User, db.Role).Values(&ids)
```
> The slice needs to be size even, every pair is for both id

> If the ids are of diffrent type, use a slice of type any
## Update
## Delete