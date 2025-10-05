MIGRATIONS_DIR = migrations

.PHONY: create-migration

create-migration:
	goose create -dir $(MIGRATIONS_DIR) -s  create_inbox_table.sql sql