build:
	docker build -t forum-app .

start:
	docker run -p 5000:5000 -p 5433:5432 --name forum-container forum-app

clear:
	docker rm -f forum-container