go get -u -t
docker compose up -d
sleep 1
go test ../ . -v -race
docker compose down