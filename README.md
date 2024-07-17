# GOE
 A SQL Like code first ORM for Go (golang)


[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/lauro-ss/goe)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

![](goe.png)
*GOE logo by [Luanexs](https://www.instagram.com/luanexs/)*

## Content
- [Install](#install)
- [Available Drivers](#available-drivers)
    - [PostgreSQL](#postgresql-🐘)
- [How to Use](#how-to-use)
    - [Quick Start](#quick-start-⚡)
    - [Database](#database-📦)
    - [Select](#select-🏷️)
        - [Select From](#select-from)
        - [Select Specific Fields](#select-specific-fields)
        - [Select Where](#select-where)
        - [Select Join](#select-join)
        - [Pagination](#pagination)
    - [Insert](#insert-🔑)
    - [Update](#update-✏️)
    - [Delete](#delete-❌)

## Install
```
go get github.com/lauro-ss/goe
```
> As any database/sql support in go, you have to get a specific driver for your database, check [Available Drivers](#available-drivers)

## Available Drivers
### PostgreSQL 🐘
```
go get github.com/lauro-ss/postgres
```
## How to Use
### Quick Start ⚡
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

	// Auto migrate all the database tables and print SQL
	db.Migrate(goe.MigrateFrom(db))

	var a Animal
	_, err = db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, 1)).Scan(&a)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fmt.Println("A error occurred when select animal", err)
		return
	}
	if a.Id == 0 {
		a.Name = "Beaver"
		a.Emoji = "🦫"
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
### Database 📦
```
// By default field "Id" is primary key
// By default all integers types are auto incremeted
type User struct {
	Id       uint
	Name     string
	Password string
	Roles    []Role `goe:"table:UserRole"`
	UserLogs []UserLog
}

// Many to one relantionship with User
// IdUser points to the user id in User table
type UserLog struct {
	Id       uint
	Action   string
	DateTime time.Time
	IdUser   User
}

// For now, in a many to many relationship
// it's necessary specify the table in goe tag
type Role struct {
	Id    uint
	Name  string
	Users []User `goe:"table:UserRole"`
}

// In goe, it's necessary to define a Database struct
// the Database struct implements *goe.DB and all
// the structs that's it's to be mappend
//
// It's through the Database struct that you will
// interact with your database
type Database struct {
	User    *User
	Role    *Role
	UserLog *UserLog
	*goe.DB
}
```
### Select 🏷️
#### Select From
```
db.Select(db.Animal).Scan(&a)
```
#### Select Specific Fields
```
db.Select(&db.Animal.Name, &db.Animal.Emoji).Scan(&a)
```
#### Select Where
```
db.Select(db.Animal).Where(db.Equals(&db.Animal.Id, 1)).Scan(&a)
```
#### Select Join
```
db.Select(db.Food).Join(db.Animal, db.Food).
    Where(db.Equals(&db.Animal.Id, 1)).Scan(&a)
```
> Join will throw an error if the tables don't have a many to many or many to one relationship.

> Using join you can access Animal Id for getting foods by Animal.
#### Pagination
```
db.Select(db.Animal).Page(2, 20).OrderByAsc(&db.Animal.Name).Scan(&animals)
```
> Get 20 habitats from second page ordered by name

same result with
```
db.Select(db.Animal).Skip(20).Take(20).OrderByAsc(&db.Animal.Name).Scan(&animals)
```
### Insert 🔑
#### Insert Table
```
var a Animal
a.Name = "Beaver"
a.Emoji = "🦫"
db.Insert(db.Animal).Value(&a)
```
#### Insert With Foreign Key
#### Insert Many To Many Table
```
```
### Update ✏️
### Delete ❌