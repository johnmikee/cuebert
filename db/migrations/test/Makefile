DATABASE:=""
NAME:=""
PASSWORD:=""
PORT:=""
USER:=""
TABLE:=""

migrations-test:
	@echo "Running migrations tests..."
	@docker run --name $(NAME) -e POSTGRES_USER=$(USER) -e POSTGRES_PASSWORD=$(PASSWORD) -e POSTGRES_DB=$(DATABASE) -p $(PORT):5432 -d postgres
	@sleep 5
	@go test -v .
	@docker stop $(NAME)
	@docker rm $(NAME)
