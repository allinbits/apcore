# Ensure data/certs directory exists
data/certs:
	mkdir -p data/certs

# Generate the key and certificate files
data/certs/key.pem data/certs/cert.pem: | data/certs
	openssl req -x509 -newkey rsa:4096 \
		-keyout data/certs/key.pem \
		-out data/certs/cert.pem \
		-days 365 -nodes \
		-subj "/CN=localhost"

.PHONY: certs down up
certs: data/certs/key.pem data/certs/cert.pem

up:
	@docker compose up -d

down:
	@docker compose down

# Remove only certificate files
clean-certs:
	@rm -rf data/certs

# Remove database files
clean-db: down
	rm -rf data/postgres

# Remove all data
clean-all: clean-certs clean-db

config.ini: down up
	@mkdir -p data/config && \
	mkdir -p data/static && \
	cd data/config && \
	go run ../../example configure

init-db: down up
	@sleep 1 && cd data/config && \
	go run ../../example init-db

init-admin: down up
	@sleep 1 && cd data/config && \
	go run ../../example init-admin

run:
	@cd data/config && \
	go run ../../example serve
