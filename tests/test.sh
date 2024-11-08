go get -u -t
docker compose up -d
sleep 2
go test ../ . -v -race
docker compose down