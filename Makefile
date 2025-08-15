# Final Complete Makefile for Go Clean Gin with Laravel-style Commands
.PHONY: build run dev test clean docker-build docker-run help install setup
.PHONY: artisan make-migration make-seeder make-entity make-package make-model
.PHONY: migrate migrate-rollback migrate-status migrate-fresh db-seed db-seed-list db-seed-specific build-artisan
.PHONY: add-column drop-column add-index db-create db-drop db-reset db-info
.PHONY: list-migrations validate-migrations init-migrations examples
.PHONY: db-mysql db-postgres db-sqlite test-all-db

# Variables
APP_NAME=go-starter
DOCKER_IMAGE=$(APP_NAME):latest
SERVER_PORT?=8080
# DB_DRIVER?=mysql

# Artisan CLI command
# ARTISAN_CMD := go run cmd/artisan/main.go
ARTISAN_CMD := $(if $(wildcard bin/artisan), $(if $(DB_DRIVER),DB_DRIVER=$(DB_DRIVER),) ./bin/artisan, $(if $(DB_DRIVER),DB_DRIVER=$(DB_DRIVER),) go run cmd/artisan/main.go)

# Default target
.DEFAULT_GOAL := help

# =============================================================================
# Basic Development Commands
# =============================================================================

## Install dependencies
install:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

## Install development tools
install-tools:
	@echo "🔧 Installing development tools..."
	@go install github.com/githubnemo/CompileDaemon@latest || echo "CompileDaemon installation failed"
	@go install github.com/air-verse/air@latest || go install github.com/cosmtrek/air@v1.49.0 || echo "Air installation failed"
	@echo "✅ Development tools installed"

## Setup project (first time)
setup: install install-tools
	@echo "🏗️  Setting up project..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "📝 Created .env file. Please configure it."; \
	fi
	@mkdir -p tmp logs bin internal/migrations internal/seeders internal/entity
	@echo "✅ Project setup complete! Run 'make dev' to start development."

## Check if port is available
check-port:
	@PORT=$${SERVER_PORT:-$(SERVER_PORT)}; \
	if lsof -i :$$PORT >/dev/null 2>&1; then \
		echo "❌ Port $$PORT is already in use"; \
		echo "Processes using port $$PORT:"; \
		lsof -i :$$PORT; \
		echo "Run 'make kill-port' to free the port"; \
		exit 1; \
	else \
		echo "✅ Port $$PORT is available"; \
	fi

## Kill process using the configured port
kill-port:
	@PORT=$${SERVER_PORT:-$(SERVER_PORT)}; \
	echo "💀 Killing processes on port $$PORT..."; \
	sudo lsof -t -i:$$PORT | xargs kill -9 2>/dev/null || echo "No processes found on port $$PORT"

## Run the application with hot reload
dev: check-port
	@if [ -f "$(shell go env GOPATH)/bin/air" ]; then \
		echo "🔥 Using Air for hot reload..."; \
		if [ ! -f .air.toml ]; then $(shell go env GOPATH)/bin/air init; fi; \
		$(shell go env GOPATH)/bin/air; \
	elif command -v CompileDaemon >/dev/null 2>&1; then \
		echo "🔥 Using CompileDaemon for hot reload..."; \
		CompileDaemon -command="./$(APP_NAME)" -build="go build -o $(APP_NAME) cmd/main.go"; \
	else \
		echo "⚡ No hot reload available, running normally..."; \
		go run cmd/main.go; \
	fi

## Force run (kill port first)
dev-force: kill-port dev

## Run without hot reload
run:
	@echo "🚀 Running application $(if $(DB_DRIVER),with DB_DRIVER=$(DB_DRIVER) database...,) "
	$(if $(DB_DRIVER),DB_DRIVER=$(DB_DRIVER),) go run cmd/main.go

## Build the application
build:
	@echo "🔨 Building application..."
	@mkdir -p bin
	go build -o bin/$(APP_NAME) cmd/main.go

## Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

## Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📋 Coverage report generated: coverage.html"
	go tool cover -func=coverage.out

## Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out coverage.html
	rm -f *.log

## Format code
fmt:
	@echo "💅 Formatting code..."
	go fmt ./...

## Tidy dependencies
tidy:
	@echo "📚 Tidying dependencies..."
	go mod tidy

# =============================================================================
# Laravel-style Artisan Commands
# =============================================================================

## Build artisan CLI tool
build-artisan:
	@echo "🎨 Building artisan CLI..."
	@mkdir -p bin
	@go build -o bin/artisan cmd/artisan/main.go
	@echo "✅ Artisan CLI built successfully"

## Create new migration file
make-migration:
	@if [ -z "$(NAME)" ] || [ -z "$(TABLE)" ]; then \
		echo "❌ Error: NAME and TABLE are required"; \
		echo "Usage: make make-migration NAME=migration_name TABLE=table_name [CREATE=true] [FIELDS=\"field1:type1,field2:type2\"] [STRATEGY=int|uuid|dual]"; \
		echo ""; \
		echo "🔑 Primary Key Strategies:"; \
		echo "  int   - ID int (primary key) - Default"; \
		echo "  uuid  - UUID uuid.UUID (primary key)"; \
		echo "  dual  - ID int (primary) + UUID (public)"; \
		echo ""; \
		echo "🗄️  Multi-Database Support:"; \
		echo "  Automatically detects DB_DRIVER (mysql|postgresql|sqlite)"; \
		echo "  Generates database-specific GORM tags"; \
		echo ""; \
		echo "Examples:"; \
		echo "  # Basic table creation"; \
		echo "  make make-migration NAME=create_users_table CREATE=true TABLE=users FIELDS=\"name:string,email:string\""; \
		echo "  # With UUID primary key"; \
		echo "  make make-migration NAME=create_products_table CREATE=true TABLE=products STRATEGY=uuid FIELDS=\"name:string,price:decimal\""; \
		echo "  # With dual strategy (int + UUID)"; \
		echo "  make make-migration NAME=create_posts_table CREATE=true TABLE=posts STRATEGY=dual FIELDS=\"title:string,content:text\""; \
		echo "  # Add column to existing table"; \
		echo "  make make-migration NAME=add_phone_to_users TABLE=users FIELDS=\"phone:string\""; \
		echo "  # Multi-database specific"; \
		echo "  DB_DRIVER=sqlite make make-migration NAME=create_categories_table CREATE=true TABLE=categories"; \
		exit 1; \
	fi
	@echo "📝 Creating migration: $(NAME)"
	@$(ARTISAN_CMD) -action=make:migration -name="$(NAME)" \
		$(if $(CREATE),-create) \
		$(if $(TABLE),-table="$(TABLE)") \
		$(if $(FIELDS),-fields="$(FIELDS)") \
		$(if $(STRATEGY),-strategy="$(STRATEGY)")

## Create new seeder file
make-seeder:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: NAME is required"; \
		echo "Usage: make make-seeder NAME=SeederName [TABLE=table_name] [DEPS=\"Seeder1,Seeder2\"]"; \
		echo ""; \
		echo "Examples:"; \
		echo "  make make-seeder NAME=UserSeeder TABLE=users"; \
		echo "  make make-seeder NAME=ProductSeeder TABLE=products DEPS=\"UserSeeder\""; \
		echo "  make make-seeder NAME=OrderSeeder DEPS=\"UserSeeder,ProductSeeder\""; \
		exit 1; \
	fi
	@echo "🌱 Creating seeder: $(NAME)"
	@$(ARTISAN_CMD) -action=make:seeder -name="$(NAME)" \
		$(if $(TABLE),-table="$(TABLE)") \
		$(if $(DEPS),-deps="$(DEPS)")

## Create new entity/model file
make-entity:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: NAME is required"; \
		echo "Usage: make make-entity NAME=ModelName [TABLE=table_name] [FIELDS=\"field1:type1|index|fk:table,field2:type2\"] [STRATEGY=int|uuid|dual]"; \
		echo ""; \
		echo "🔑 Primary Key Strategies:"; \
		echo "  int   - ID int (primary key, auto-increment) - Default"; \
		echo "  uuid  - UUID uuid.UUID (primary key, auto-generated)"; \
		echo "  dual  - ID int (primary, relations) + UUID (public API)"; \
		echo ""; \
		echo "Examples:"; \
		echo "  # Basic entity with int ID"; \
		echo "  make make-entity NAME=Category FIELDS=\"name:string,description:text\""; \
		echo "  # Entity with UUID primary key"; \
		echo "  make make-entity NAME=Product STRATEGY=uuid FIELDS=\"name:string,price:decimal\""; \
		echo "  # Entity with dual strategy (recommended for users)"; \
		echo "  make make-entity NAME=User STRATEGY=dual FIELDS=\"name:string,email:string\""; \
		echo "  # With custom table and indexes"; \
		echo "  make make-entity NAME=Post TABLE=tb_posts STRATEGY=dual FIELDS=\"title:string|index,user_id:uuid|fk:tb_users\""; \
		exit 1; \
	fi
	@echo "📋 Creating entity: $(NAME) with $(or $(STRATEGY),int) strategy"
	@$(ARTISAN_CMD) -action=make:model -name="$(NAME)" \
		-table="$(or $(TABLE),$(shell echo $(NAME) | tr '[:upper:]' '[:lower:]')s)" \
		$(if $(FIELDS),-fields="$(FIELDS)") \
		$(if $(STRATEGY),-strategy="$(STRATEGY)")

## Create new package with handler, usecase, repository, port
make-package:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: NAME is required"; \
		echo "Usage: make make-package NAME=PackageName"; \
		echo ""; \
		echo "Example:"; \
		echo "  make make-package NAME=Product"; \
		exit 1; \
	fi
	@echo "📦 Creating package: $(NAME)"
	@$(ARTISAN_CMD) -action=make:package -name="$(NAME)"

## Create model with migration and seeder (complete stack)
make-model:
	@if [ -z "$(NAME)" ] || [ -z "$(TABLE)" ]; then \
		echo "❌ Error: NAME and TABLE are required"; \
		echo "Usage: make make-model NAME=ModelName TABLE=table_name [FIELDS=\"field1:type1,field2:type2\"] [STRATEGY=int|uuid|dual]"; \
		echo ""; \
		echo "🔑 Primary Key Strategies:"; \
		echo "  int   - ID int (primary) - Best for internal systems"; \
		echo "  uuid  - UUID (primary) - Best for distributed systems"; \
		echo "  dual  - ID int (primary) + UUID (public) - Best for user-facing APIs"; \
		echo ""; \
		echo "🗄️  Multi-Database Support:"; \
		echo "  Automatically generates correct GORM tags for your DB_DRIVER"; \
		echo "  Supports MySQL, PostgreSQL, SQLite"; \
		echo ""; \
		echo "Examples:"; \
		echo "  # Basic model with int ID (default)"; \
		echo "  make make-model NAME=Category TABLE=categories FIELDS=\"name:string,description:text\""; \
		echo "  # User model with dual strategy (recommended)"; \
		echo "  make make-model NAME=User TABLE=users STRATEGY=dual FIELDS=\"name:string,email:string,age:int\""; \
		echo "  # Product model with UUID primary key"; \
		echo "  make make-model NAME=Product TABLE=products STRATEGY=uuid FIELDS=\"name:string,price:decimal,sku:string\""; \
		echo "  # Multi-database example"; \
		echo "  DB_DRIVER=sqlite make make-model NAME=Post TABLE=posts STRATEGY=dual FIELDS=\"title:string,content:text\""; \
		exit 1; \
	fi
	@echo "🏗️  Creating complete model stack for: $(NAME) with $(or $(STRATEGY),int) strategy"
	@echo "📋 Step 1: Creating entity struct..."
	@$(ARTISAN_CMD) -action=make:model -name="$(NAME)" \
		$(if $(TABLE),-table="$(TABLE)") \
		$(if $(FIELDS),-fields="$(FIELDS)") \
		$(if $(STRATEGY),-strategy="$(STRATEGY)")
	@echo "📄 Step 2: Creating migration (without entity)..."
	@$(ARTISAN_CMD) -action=make:migration -name="create_$(shell echo $(NAME) | tr '[:upper:]' '[:lower:]')_table" \
		-create -table="$(TABLE)" -skip-entity \
		$(if $(FIELDS),-fields="$(FIELDS)") \
		$(if $(STRATEGY),-strategy="$(STRATEGY)")
	@echo "🌱 Step 3: Creating seeder..."
	@$(MAKE) make-seeder NAME=$(NAME)Seeder TABLE=$(TABLE)
	@echo "✅ Complete model stack created successfully!"
	@echo "📁 Files created:"
	@echo "  - internal/entity/$(shell echo $(NAME) | tr '[:upper:]' '[:lower:]').go (Entity struct)"
	@echo "  - internal/migrations/TIMESTAMP_create_$(shell echo $(NAME) | tr '[:upper:]' '[:lower:]')s_table.go (Migration - no duplicate entity)"
	@echo "  - internal/seeders/$(shell echo $(NAME) | tr '[:upper:]' '[:lower:]')_seeder.go (Seeder)"

# =============================================================================
# Migration Management Commands
# =============================================================================

## Run pending migrations
migrate:
	@echo "⬆️  Running migrations $(if $(DB_DRIVER),with DB_DRIVER=$(DB_DRIVER) database...,) "
	@$(ARTISAN_CMD) -action=migrate

## Rollback migrations
migrate-rollback:
	@echo "⬇️  Rolling back migrations..."
	@$(ARTISAN_CMD) -action=migrate:rollback \
		$(if $(COUNT),-count=$(COUNT))

## Show migration status
migrate-status:
	@echo "📊 Checking migration status..."
	@$(ARTISAN_CMD) -action=migrate:status

## Fresh migration (DANGER!)
migrate-fresh:
	@echo "🚨 WARNING: This will destroy all data!"
	@read -p "Type 'FRESH' to continue: " -r; \
	if [ "$$REPLY" = "FRESH" ]; then \
		echo "🗑️  Dropping all tables..."; \
		$(MAKE) migrate-rollback COUNT=all; \
		echo "⬆️  Running fresh migrations..."; \
		$(MAKE) migrate; \
		echo "🌱 Running seeders..."; \
		$(MAKE) db-seed; \
		echo "✅ Fresh migration completed!"; \
	else \
		echo "❌ Cancelled"; \
	fi

## Run database seeders
db-seed:
	@echo "🌱 Running seeders with dependency resolution..."
	@$(ARTISAN_CMD) -action=db:seed $(if $(NAME),-name=$(NAME))

## List all seeders with their dependencies
db-seed-list:
	@echo "📋 Listing all registered seeders with dependencies..."
	@$(ARTISAN_CMD) -action=db:seed -name=list

## Run specific seeder with its dependencies
db-seed-specific:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: NAME is required"; \
		echo "Usage: make db-seed-specific NAME=SeederName"; \
		echo ""; \
		echo "Example:"; \
		echo "  make db-seed-specific NAME=ProductSeeder"; \
		echo "  # This will run UserSeeder first, then ProductSeeder"; \
		exit 1; \
	fi
	@echo "🌱 Running seeder: $(NAME) (with dependencies)"
	@$(ARTISAN_CMD) -action=db:seed -name=$(NAME)

# =============================================================================
# Laravel-style Shortcuts for Common Operations
# =============================================================================

## Add column to existing table (TABLE=users COLUMN=phone TYPE=string)
add-column:
	@if [ -z "$(TABLE)" ] || [ -z "$(COLUMN)" ] || [ -z "$(TYPE)" ]; then \
		echo "❌ Error: TABLE, COLUMN, and TYPE are required"; \
		echo "Usage: make add-column TABLE=table_name COLUMN=column_name TYPE=column_type"; \
		echo ""; \
		echo "Example:"; \
		echo "  make add-column TABLE=users COLUMN=phone TYPE=string"; \
		exit 1; \
	fi
	@$(MAKE) make-migration NAME=add_$(COLUMN)_to_$(TABLE) TABLE=$(TABLE) FIELDS="$(COLUMN):$(TYPE)"

## Drop column from table (TABLE=users COLUMN=phone)
drop-column:
	@if [ -z "$(TABLE)" ] || [ -z "$(COLUMN)" ]; then \
		echo "❌ Error: TABLE and COLUMN are required"; \
		echo "Usage: make drop-column TABLE=table_name COLUMN=column_name"; \
		echo ""; \
		echo "Example:"; \
		echo "  make drop-column TABLE=users COLUMN=old_field"; \
		exit 1; \
	fi
	@$(MAKE) make-migration NAME=drop_$(COLUMN)_from_$(TABLE)

## Add index to table (TABLE=products COLUMNS="category,price")
add-index:
	@if [ -z "$(TABLE)" ] || [ -z "$(COLUMNS)" ]; then \
		echo "❌ Error: TABLE and COLUMNS are required"; \
		echo "Usage: make add-index TABLE=table_name COLUMNS=\"col1,col2\""; \
		echo ""; \
		echo "Example:"; \
		echo "  make add-index TABLE=products COLUMNS=\"category,price\""; \
		exit 1; \
	fi
	@$(MAKE) make-migration NAME=add_index_to_$(TABLE)_on_$(shell echo $(COLUMNS) | tr ',' '_')

# =============================================================================
# Database Management Commands
# =============================================================================

## Create database
db-create:
	@echo "🏗️  Creating database..."
	@PGPASSWORD=$(DB_PASSWORD) createdb -h $(DB_HOST) -U $(DB_USER) $(DB_NAME) 2>/dev/null || echo "Database might already exist"

## Drop database (DANGER!)
db-drop:
	@echo "🚨 WARNING: This will drop the entire database!"
	@read -p "Type 'DROP' to continue: " -r; \
	if [ "$$REPLY" = "DROP" ]; then \
		PGPASSWORD=$(DB_PASSWORD) dropdb -h $(DB_HOST) -U $(DB_USER) $(DB_NAME) 2>/dev/null || echo "Database might not exist"; \
		echo "✅ Database dropped"; \
	else \
		echo "❌ Cancelled"; \
	fi

## Reset database completely
db-reset: db-drop db-create migrate db-seed

## Show database info
db-info:
	@echo "📊 Database Information:"
	@echo "Current DB Type: $(if $(DB_DRIVER),$(DB_DRIVER),mysql)"
	@if [ -f .env ]; then \
		source .env; \
		echo "Host: $$DB_HOST"; \
		echo "Port: $$DB_PORT"; \
		echo "Database: $$DB_NAME"; \
		echo "User: $$DB_USER"; \
	else \
		echo "No .env file found"; \
	fi

## Switch database type for current session
db-mysql:
	@echo "🐬 Switching to MySQL database..."
	@$(MAKE) DB_DRIVER=mysql $(filter-out db-mysql,$(MAKECMDGOALS))

## Switch database type for current session  
db-postgres:
	@echo "🐘 Switching to PostgreSQL database..."
	@$(MAKE) DB_DRIVER=postgresql $(filter-out db-postgres,$(MAKECMDGOALS))

## Switch database type for current session
db-sqlite:
	@echo "🗄️  Switching to SQLite database..."
	@$(MAKE) DB_DRIVER=sqlite $(filter-out db-sqlite,$(MAKECMDGOALS))

## Test all database types
test-all-db:
	@echo "🧪 Testing all database types..."
	@echo "Testing MySQL..."
	@$(MAKE) DB_DRIVER=mysql migrate-status || echo "MySQL test failed"
	@echo "Testing PostgreSQL..."
	@$(MAKE) DB_DRIVER=postgresql migrate-status || echo "PostgreSQL test failed"  
	@echo "Testing SQLite..."
	@$(MAKE) DB_DRIVER=sqlite migrate-status || echo "SQLite test failed"
	@echo "✅ Multi-database testing complete"

# Handle arguments for db-* targets
%:
	@:

# =============================================================================
# Development Utilities
# =============================================================================

## Create directories for migrations and seeders
init-migrations:
	@echo "📁 Creating migration directories..."
	@mkdir -p internal/migrations internal/seeders internal/entity
	@echo "✅ Migration directories created"

## List all migration files
list-migrations:
	@echo "📂 Migration files:"
	@if [ -d "internal/migrations" ]; then \
		find internal/migrations -name "*.go" -type f | sort; \
	else \
		echo "No migrations directory found"; \
	fi
	@echo ""
	@echo "📂 Seeder files:"
	@if [ -d "internal/seeders" ]; then \
		find internal/seeders -name "*.go" -type f | sort; \
	else \
		echo "No seeders directory found"; \
	fi
	@echo ""
	@echo "📂 Entity files:"
	@if [ -d "internal/entity" ]; then \
		find internal/entity -name "*.go" -type f | sort; \
	else \
		echo "No entity directory found"; \
	fi

## Validate migration files
validate-migrations:
	@echo "🔍 Validating migration files..."
	@if [ -d "internal/migrations" ]; then \
		for file in internal/migrations/*.go; do \
			if [ -f "$$file" ]; then \
				echo "Checking $$file..."; \
				go vet "$$file" || exit 1; \
			fi \
		done; \
		echo "✅ All migration files are valid"; \
	else \
		echo "No migrations directory found"; \
	fi

# =============================================================================
# Docker Commands
# =============================================================================

## Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

## Run Docker containers
docker-run:
	@echo "🐳 Starting Docker containers..."
	docker compose up -d

## Stop Docker containers
docker-stop:
	@echo "🐳 Stopping Docker containers..."
	docker compose down

## View Docker logs
docker-logs:
	@echo "📋 Showing Docker logs..."
	docker compose logs -f

# =============================================================================
# Health & Monitoring Commands
# =============================================================================

## Health check
health:
	@echo "❤️  Checking application health..."
	@curl -f http://localhost:$(SERVER_PORT)/health || echo "Health check failed"

## Show application status
status:
	@echo "📊 Application Status:"
	@echo "Server: http://localhost:$(SERVER_PORT)"
	@$(MAKE) health
	@$(MAKE) db-info

# =============================================================================
# Help & Examples
# =============================================================================

## Show usage examples
examples:
	@echo "📖 Laravel-style Command Examples with Multi-Database & Primary Key Strategies:"
	@echo ""
	@echo "🔑 Primary Key Strategy Examples:"
	@echo "  # int strategy (default) - Best for internal systems"
	@echo "  make make-model NAME=Category TABLE=categories FIELDS=\"name:string,description:text\""
	@echo "  # Result: ID int (primary key, auto-increment)"
	@echo ""
	@echo "  # uuid strategy - Best for distributed systems"
	@echo "  make make-model NAME=Product TABLE=products STRATEGY=uuid FIELDS=\"name:string,price:decimal,sku:string\""
	@echo "  # Result: UUID uuid.UUID (primary key, auto-generated)"
	@echo ""
	@echo "  # dual strategy - Best for user-facing APIs"
	@echo "  make make-model NAME=User TABLE=users STRATEGY=dual FIELDS=\"name:string,email:string,age:int\""
	@echo "  # Result: ID int (primary key for DB relations) + UUID uuid.UUID (public API identifier)"
	@echo ""
	@echo "🗄️  Multi-Database Examples:"
	@echo "  # SQLite (development)"
	@echo "  DB_DRIVER=sqlite make make-model NAME=Post TABLE=posts STRATEGY=dual FIELDS=\"title:string,content:text\""
	@echo "  # MySQL (production)"
	@echo "  DB_DRIVER=mysql make migrate"
	@echo "  # PostgreSQL (production)"
	@echo "  DB_DRIVER=postgresql make db-seed"
	@echo ""
	@echo "📦 Creating Complete Features with Strategies:"
	@echo "  # E-commerce system with proper primary key strategies"
	@echo "  make make-model NAME=User TABLE=users STRATEGY=dual FIELDS=\"name:string,email:string\"      # Users need public UUIDs"
	@echo "  make make-model NAME=Category TABLE=categories FIELDS=\"name:string,description:text\"        # Categories can use int IDs"
	@echo "  make make-model NAME=Product TABLE=products STRATEGY=uuid FIELDS=\"name:string,price:decimal\" # Products need distributed UUIDs"
	@echo "  make make-model NAME=Order TABLE=orders STRATEGY=dual FIELDS=\"total:decimal,status:string\"   # Orders need both for relations + public API"
	@echo "  make make-package NAME=Product"
	@echo "  make migrate && make db-seed"
	@echo ""
	@echo "🏗️  Creating Individual Components:"
	@echo "  # Create just entity with specific strategy"
	@echo "  make make-entity NAME=User TABLE=users STRATEGY=dual FIELDS=\"name:string,email:string,age:int\""
	@echo ""
	@echo "  # Create just package structure"
	@echo "  make make-package NAME=Product"
	@echo ""
	@echo "  # Create table migration with strategy"
	@echo "  make make-migration NAME=create_posts_table CREATE=true TABLE=posts STRATEGY=dual FIELDS=\"title:string,content:text\""
	@echo ""
	@echo "📝 Adding Columns & Indexes:"
	@echo "  make add-column TABLE=users COLUMN=phone TYPE=string"
	@echo "  make add-column TABLE=products COLUMN=sku TYPE=string"
	@echo "  make add-index TABLE=products COLUMNS=\"category,price\""
	@echo "  make drop-column TABLE=users COLUMN=old_field"
	@echo ""
	@echo "🌱 Seeding & Migration (with Dependencies):"
	@echo "  # Create seeders with dependencies"
	@echo "  make make-seeder NAME=UserSeeder TABLE=users"
	@echo "  make make-seeder NAME=ProductSeeder TABLE=products DEPS=\"UserSeeder\""
	@echo "  make make-seeder NAME=OrderSeeder DEPS=\"UserSeeder,ProductSeeder\""
	@echo ""
	@echo "  # Run seeders (automatic dependency resolution)"
	@echo "  make db-seed                   # Run all seeders in correct order"
	@echo "  make db-seed-list              # Show all seeders with dependencies"
	@echo "  make db-seed-specific NAME=ProductSeeder  # Run ProductSeeder (+ UserSeeder first)"
	@echo ""
	@echo "  # Migration management"
	@echo "  make migrate                   # Run pending migrations"
	@echo "  make migrate-status            # Show status"
	@echo "  make migrate-rollback          # Rollback last migration"
	@echo "  make migrate-rollback COUNT=3  # Rollback last 3 migrations"
	@echo ""
	@echo "🔄 Multi-Database Workflow:"
	@echo "  # Development with SQLite"
	@echo "  DB_DRIVER=sqlite make migrate"
	@echo "  DB_DRIVER=sqlite make db-seed"
	@echo "  DB_DRIVER=sqlite make dev"
	@echo ""
	@echo "  # Production with MySQL"
	@echo "  DB_DRIVER=mysql make migrate"
	@echo "  DB_DRIVER=mysql make db-seed"
	@echo ""
	@echo "  # Test all databases"
	@echo "  make test-all-db               # Test MySQL, PostgreSQL, SQLite"
	@echo ""
	@echo "📁 Complete Workflow Example with Primary Key Strategies:"
	@echo "  # 1. Setup project"
	@echo "  make setup"
	@echo "  make build-artisan"
	@echo ""
	@echo "  # 2. Create models with appropriate strategies"
	@echo "  make make-model NAME=User TABLE=users STRATEGY=dual FIELDS=\"name:string,email:string\"         # dual for user-facing"
	@echo "  make make-model NAME=Category TABLE=categories FIELDS=\"name:string,description:text\"          # int for internal"
	@echo "  make make-model NAME=Product TABLE=products STRATEGY=uuid FIELDS=\"name:string,price:decimal\"  # uuid for distributed"
	@echo "  make make-model NAME=Order TABLE=orders STRATEGY=dual FIELDS=\"total:decimal,status:string\"    # dual for API + relations"
	@echo ""
	@echo "  # 3. Create seeders with dependencies"
	@echo "  make make-seeder NAME=UserSeeder TABLE=users"
	@echo "  make make-seeder NAME=CategorySeeder TABLE=categories"
	@echo "  make make-seeder NAME=ProductSeeder TABLE=products DEPS=\"CategorySeeder\""
	@echo "  make make-seeder NAME=OrderSeeder TABLE=orders DEPS=\"UserSeeder,ProductSeeder\""
	@echo ""
	@echo "  # 4. Deploy (seeders will run in correct order automatically)"
	@echo "  make migrate"
	@echo "  make db-seed    # UserSeeder → CategorySeeder → ProductSeeder → OrderSeeder"
	@echo "  make dev"
	@echo ""
	@echo "📊 Primary Key Strategy Decision Guide:"
	@echo "  int   - Internal systems, admin panels, configuration tables"
	@echo "  uuid  - Microservices, public APIs, distributed systems, external integrations"
	@echo "  dual  - User-facing applications where you need both performance (int) and security (UUID)"
	@echo ""
	@echo "🗄️  Database-Specific Features:"
	@echo "  SQLite    - CURRENT_TIMESTAMP (no microseconds)"
	@echo "  MySQL     - CURRENT_TIMESTAMP(3) with ON UPDATE"
	@echo "  PostgreSQL- gen_random_uuid() for UUID defaults"

## Show help with all available commands (Updated)
help:
	@echo "🚀 Go Starter - Multi-Database Laravel-style Development with Primary Key Strategies"
	@echo ""
	@echo "🏗️  Setup & Development:"
	@echo "  setup              Setup project (first time)"
	@echo "  dev                Run with hot reload"
	@echo "  dev-force          Kill port conflicts and run"
	@echo "  run                Run without hot reload"
	@echo "  build              Build application"
	@echo "  build-artisan      Build artisan CLI tool"
	@echo ""
	@echo "🎨 Laravel-style Generators (with Primary Key Strategies):"
	@echo "  make-migration     Create new migration file [STRATEGY=int|uuid|dual]"
	@echo "  make-seeder        Create seeder with dependency support"
	@echo "  make-entity        Create new entity/model file [STRATEGY=int|uuid|dual]"
	@echo "  make-package       Create new package (handler, usecase, repository, port)"
	@echo "  make-model         Create complete model stack (entity + migration + seeder) [STRATEGY=int|uuid|dual]"
	@echo ""
	@echo "🔑 Primary Key Strategies:"
	@echo "  int                ID int (primary key, auto-increment) - Default"
	@echo "  uuid               UUID uuid.UUID (primary key, auto-generated)"
	@echo "  dual               ID int (primary, relations) + UUID (public API)"
	@echo ""
	@echo "⚡ Quick Actions:"
	@echo "  add-column         Add column to existing table"
	@echo "  drop-column        Drop column from table"
	@echo "  add-index          Add index to table"
	@echo ""
	@echo "🗄️  Migration & Database:"
	@echo "  migrate            Run pending migrations"
	@echo "  migrate-status     Show migration status"
	@echo "  migrate-rollback   Rollback migrations"
	@echo "  migrate-fresh      Fresh migration (DANGER!)"
	@echo ""
	@echo "🌱 Database Seeding (with Dependencies):"
	@echo "  db-seed            Run all seeders (auto-resolves dependencies)"
	@echo "  db-seed-list       List all seeders with their dependencies"
	@echo "  db-seed-specific   Run specific seeder with its dependencies"
	@echo ""
	@echo "🏭 Database Management:"
	@echo "  db-create          Create database"
	@echo "  db-drop            Drop database (DANGER!)"
	@echo "  db-reset           Reset database completely"
	@echo "  db-info            Show database information"
	@echo ""
	@echo "🔄 Multi-Database Support:"
	@echo "  db-mysql           Switch to MySQL for commands"
	@echo "  db-postgres        Switch to PostgreSQL for commands"
	@echo "  db-sqlite          Switch to SQLite for commands"
	@echo "  test-all-db        Test all database types"
	@echo ""
	@echo "🔍 Utilities:"
	@echo "  list-migrations    List all migration/seeder/entity files"
	@echo "  validate-migrations Validate migration syntax"
	@echo "  init-migrations    Create migration directories"
	@echo "  examples           Show detailed usage examples"
	@echo ""
	@echo "🧪 Testing & Quality:"
	@echo "  test               Run tests"
	@echo "  test-coverage      Run tests with coverage"
	@echo "  fmt                Format code"
	@echo "  tidy               Tidy dependencies"
	@echo "  clean              Clean build artifacts"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  docker-build       Build Docker image"
	@echo "  docker-run         Start containers"
	@echo "  docker-stop        Stop containers"
	@echo "  docker-logs        View container logs"
	@echo ""
	@echo "❤️  Monitoring:"
	@echo "  health             Check application health"
	@echo "  status             Show application status"
	@echo ""
	@echo "💡 New Features (v2.0):"
	@echo "  🔑 Primary Key Strategies: Choose int, uuid, or dual for your entities"
	@echo "  🗄️  Multi-Database: Support MySQL, PostgreSQL, SQLite with same commands"
	@echo "  🔗 Seeder Dependencies: Seeders automatically run in correct order"
	@echo "  📊 Dependency Visualization: See which seeders depend on others"
	@echo "  🎯 Smart Execution: Run specific seeder with auto-dependency resolution"
	@echo "  🏗️  Dynamic Migration Discovery: No manual imports needed"
	@echo "  🛡️  Error Helper Functions: Simplified error handling"
	@echo "  🏭 Container Design: Factory pattern with interface-based DI"
	@echo ""
	@echo "🌟 Quick Start Examples:"
	@echo "  # Create user model with dual strategy (recommended)"
	@echo "  make make-model NAME=User TABLE=users STRATEGY=dual FIELDS=\"name:string,email:string\""
	@echo ""
	@echo "  # Create product model with UUID strategy"
	@echo "  make make-model NAME=Product TABLE=products STRATEGY=uuid FIELDS=\"name:string,price:decimal\""
	@echo ""
	@echo "  # Run migrations on different databases"
	@echo "  make db-sqlite migrate        # SQLite (development)"
	@echo "  make db-mysql migrate         # MySQL (production)"
	@echo "  make db-postgres migrate      # PostgreSQL (production)"
	@echo ""
	@echo "  # Environment variable approach"
	@echo "  DB_DRIVER=sqlite make migrate   # SQLite"
	@echo "  DB_DRIVER=mysql make db-seed    # MySQL"
	@echo ""
	@echo "  # Test all databases"
	@echo "  make test-all-db              # Test MySQL, PostgreSQL, SQLite"
	@echo ""
	@echo "📖 For detailed examples: make examples"
	@echo "🔗 Laravel-style workflow: https://laravel.com/docs/migrations"
	@echo "📚 Documentation: Check README.md for complete guide"

# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif