module github.com/olauro/goe/tests

go 1.22.1

require (
	github.com/google/uuid v1.6.0
	github.com/olauro/goe v0.2.2
	github.com/olauro/postgres v0.2.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.29.0 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/text v0.20.0 // indirect
)

replace github.com/olauro/goe => ../
