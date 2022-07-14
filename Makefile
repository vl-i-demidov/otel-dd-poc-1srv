.PHONY: restart
restart:
	docker compose down --remove-orphans && docker compose build && DD_API_KEY=$(DD_API_KEY) docker compose up -d

.PHONY: stop
stop:
	docker compose down --remove-orphans

.PHONY: test
test:
	curl http://localhost:8080/?sleep=$(SLEEP)