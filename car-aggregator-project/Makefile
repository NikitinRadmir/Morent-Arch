.PHONY: all docker greetings clean env

all: env docker

docker: greetings
	docker compose up -d

docker-down: 
	docker compose down

env:
	@if [ ! -f .env ]; then \
		echo "Создание файла .env из .env.example..."; \
		cp .env.example .env; \
		echo "Файл .env создан. Пожалуйста, откройте его и заполните необходимые секретные ключи."; \
		echo "!!! ВНИМАНИЕ: Запуск остановлен. Пожалуйста, отредактируйте .env и запустите 'make' снова."; \
		exit 1; \
	fi
	@if grep -q '<ЗАПОЛНИТЕ' .env; then \
		echo "!!! ВНИМАНИЕ: В файле .env все еще есть незаполненные секретные ключи."; \
		echo "Пожалуйста, отредактируйте .env и замените все значения, содержащие '<ЗАПОЛНИТЕ...>'."; \
		echo "Запуск остановлен."; \
		exit 1; \
	fi


greetings:
	@echo "Hello! We will start the docker container for you. Please wait a moment..."

clean:
	docker compose down