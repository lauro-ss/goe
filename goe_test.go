package goe_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lauro-ss/goe"
)

type MockDriver struct {
}

func (md *MockDriver) Migrate(*goe.Migrator, goe.Connection) {
}

func (md *MockDriver) Init(*goe.DB) {
}

func (md *MockDriver) KeywordHandler(s string) string {
	return fmt.Sprintf(`"%s"`, s)
}

func TestMapDatabase(t *testing.T) {
	type User struct {
		Id   uint
		Name string
	}

	type UserLog struct {
		Id     uint
		Action string
		IdUser uint `goe:"table:User"`
	}

	type Database struct {
		User    *User
		UserLog *UserLog
		*goe.DB
	}

	db := &Database{DB: &goe.DB{}}
	err := goe.Open(db, &MockDriver{}, goe.Config{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMapDatabaseErrorPrimaryKey(t *testing.T) {
	type User struct {
		IdUser string
		Name   string
	}

	type Database struct {
		User *User
		*goe.DB
	}

	db := &Database{DB: &goe.DB{}}
	err := goe.Open(db, &MockDriver{}, goe.Config{})
	if !errors.Is(err, goe.ErrStructWithoutPrimaryKey) {
		t.Fatal("Was expected a goe.ErrStructWithoutPrimaryKey but get:", err)
	}
}

func TestMapDatabaseErrorManyToOne(t *testing.T) {
	type UserLog struct {
		Id     uint
		Action string
		IdUser uint `goe:"table:UserRole"`
	}

	type User struct {
		Id   uint
		Name string
		Logs []UserLog
	}

	type Database struct {
		User    *User
		UserLog *UserLog
		*goe.DB
	}

	db := &Database{DB: &goe.DB{}}
	err := goe.Open(db, &MockDriver{}, goe.Config{})
	if !errors.Is(err, goe.ErrInvalidManyToOne) {
		t.Fatal("Was expected a goe.ErrInvalidManyToOne but get:", err)
	}
}

func TestMapDatabaseErrorOneToOne(t *testing.T) {
	type UserLog struct {
		Id     uint
		Action string
		IdUser uint `goe:"table:UserRole"`
	}

	type User struct {
		Id   uint
		Name string
	}

	type Database struct {
		User    *User
		UserLog *UserLog
		*goe.DB
	}

	db := &Database{DB: &goe.DB{}}
	err := goe.Open(db, &MockDriver{}, goe.Config{})
	if !errors.Is(err, goe.ErrInvalidManyToOne) {
		t.Fatal("Was expected a goe.ErrInvalidManyToOne but get:", err)
	}
}
