package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/lauro-ss/goe"
)

type Driver struct {
	dns string
}

func Open(dns string) (driver *Driver) {
	return &Driver{dns: dns}
}

func (dr *Driver) Init(db *goe.DB) {
	if db.ConnPool != nil {
		return
	}
	config, err := pgx.ParseConfig(dr.dns)
	if err != nil {
		//TODO: Add error handling
		fmt.Println(err)
		return
	}
	db.ConnPool = stdlib.OpenDB(*config)
}

func (dr *Driver) KeywordHandler(s string) string {
	return fmt.Sprintf(`"%s"`, s)
}
