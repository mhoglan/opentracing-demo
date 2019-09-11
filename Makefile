
up:
	docker-compose up -d

down:
	docker-compose down

services:
	./create_services.sh ping.json  http://localhost:8181
	./create_services.sh echo.json http://localhost:8181
	./create_services.sh pong.json http://localhost:8182

ping:
	curl http://localhost:8181/ping
