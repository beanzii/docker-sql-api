server:
	go run main.go .env

build:
	go build -o /bin/server main.go

d.up:
	docker-compose up

d.up.bg:
	docker-compose up -d

d.down:
	docker-compose down

d.build:
	docker-compose build

d.prune:
	docker system prune -a

d.prune.vol:
	docker system prune -a --volumes
