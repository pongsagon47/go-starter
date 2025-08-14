// cmd/artisan/main.go - Complete Laravel-style CLI tool
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"go-starter/config"
	pkgDatabase "go-starter/pkg/database"
	"go-starter/pkg/logger"

	// Import to register migrations and seeders
	_ "go-starter/internal/migrations"
	_ "go-starter/internal/seeders"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	action = flag.String("action", "", "Action: make:migration, make:seeder, make:model, make:package, migrate, migrate:rollback, migrate:status")
	name   = flag.String("name", "", "Migration/Seeder/Model/Package name")
	table  = flag.String("table", "", "Table name for migration")
	create = flag.Bool("create", false, "Create table migration")
	fields = flag.String("fields", "", "Fields for migration (name:type,email:string)")
	deps   = flag.String("deps", "", "Dependencies for seeder (UserSeeder,CategorySeeder)") // เพิ่มบรรทัดนี้
	count  = flag.String("count", "1", "Number of migrations to rollback")
	help   = flag.Bool("help", false, "Show help")
)

func main() {
	flag.Parse()

	if *help || *action == "" {
		showHelp()
		return
	}

	switch *action {
	case "make:migration":
		if *name == "" || *table == "" {
			fmt.Println("❌ Migration name is required")
			fmt.Println("Usage: go run cmd/artisan/main.go -action=make:migration -name=migration_name -table=table_name")
			os.Exit(1)
		}
		createMigration(*name, *table, *create, *fields)

	case "make:seeder":
		if *name == "" {
			fmt.Println("❌ Seeder name is required")
			fmt.Println("Usage: go run cmd/artisan/main.go -action=make:seeder -name=seeder_name")
			os.Exit(1)
		}
		createSeeder(*name, *table, *deps)

	case "make:model":
		if *name == "" || *table == "" {
			fmt.Println("❌ Model name is required")
			fmt.Println("Usage: go run cmd/artisan/main.go -action=make:model -name=model_name -table=table_name")
			os.Exit(1)
		}
		createModel(*name, *table, *fields)

	case "make:package":
		if *name == "" {
			fmt.Println("❌ Package name is required")
			fmt.Println("Usage: go run cmd/artisan/main.go -action=make:package -name=package_name")
			os.Exit(1)
		}
		createPackage(*name)

	case "migrate":
		runMigrations()

	case "migrate:rollback":
		rollbackMigrations(*count)

	case "migrate:status":
		showMigrationStatus()

	case "db:seed":
		runSeeders(*name)

	default:
		fmt.Printf("❌ Unknown action: %s\n", *action)
		showHelp()
		os.Exit(1)
	}
}

// createMigration function in main.go
func createMigration(migrationName, tableName string, isCreate bool, fieldList string) {
	timestamp := time.Now().Format("2006_01_02_150405")
	fileName := fmt.Sprintf("%s_%s.go", timestamp, toSnakeCase(migrationName))

	// Create migrations directory if not exists
	migrationsDir := "internal/migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create migrations directory: %v\n", err)
		os.Exit(1)
	}

	filePath := filepath.Join(migrationsDir, fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("❌ Migration file already exists: %s\n", filePath)
		os.Exit(1)
	}

	// Use the new parseFields function
	parsedFields := parseFields(fieldList)

	// Create migration data
	data := MigrationData{
		ClassName:   toPascalCase(migrationName),
		TableName:   tableName,
		Timestamp:   timestamp,
		Description: migrationName,
		Fields:      parsedFields,
		Version:     fmt.Sprintf("%s_%s", timestamp, migrationName),
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("❌ Failed to create migration file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Choose template
	var tmpl *template.Template
	if isCreate && tableName != "" {
		tmpl = template.Must(template.New("create_table").Funcs(templateFuncs).Parse(createTableTemplate))
	} else if tableName != "" {
		tmpl = template.Must(template.New("alter_table").Funcs(templateFuncs).Parse(alterTableTemplate))
	} else {
		tmpl = template.Must(template.New("migration").Funcs(templateFuncs).Parse(migrationTemplate))
	}

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		fmt.Printf("❌ Failed to generate migration file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Migration created: %s\n", filePath)
	fmt.Printf("📝 Class: %s\n", data.ClassName)
	if tableName != "" {
		fmt.Printf("🗂️  Table: %s\n", tableName)
	}

	// Show field summary
	if len(parsedFields) > 0 {
		fmt.Printf("📋 Fields:\n")
		for _, field := range parsedFields {
			extras := []string{}
			if field.HasIndex {
				extras = append(extras, "indexed")
			}
			if field.IsForeignKey {
				extras = append(extras, fmt.Sprintf("FK->%s", field.FKReference))
			}

			extraStr := ""
			if len(extras) > 0 {
				extraStr = fmt.Sprintf(" (%s)", strings.Join(extras, ", "))
			}

			fmt.Printf("  - %s: %s%s\n", field.Name, field.Type, extraStr)
		}
	}

	// Auto-create entity if this is a create table migration
	if isCreate && tableName != "" {
		fmt.Printf("\n🚀 Auto-creating entity...\n")
		if err := autoCreateEntity(tableName, parsedFields); err != nil {
			fmt.Printf("⚠️  Warning: Failed to create entity: %v\n", err)
		}
	}
}

func autoCreateEntity(tableName string, fields []Field) error {
	// Generate entity name from table name
	entityName := getStructName(tableName)
	// fileName := fmt.Sprintf("%s.go", strings.ToLower(entityName))
	fileName := fmt.Sprintf("%s.go", toSnakeCase(entityName))

	// Create entity directory if not exists
	entityDir := "internal/entity"
	if err := os.MkdirAll(entityDir, 0755); err != nil {
		return fmt.Errorf("failed to create entity directory: %w", err)
	}

	filePath := filepath.Join(entityDir, fileName)

	// Check if file already exists - warn but don't fail
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("⚠️  Entity file already exists, skipping: %s\n", filePath)
		return nil
	}

	// Create entity data
	data := EntityData{
		EntityName: entityName,
		TableName:  tableName,
		Fields:     fields,
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create entity file: %w", err)
	}
	defer file.Close()

	// Execute template
	tmpl := template.Must(template.New("entity").Funcs(templateFuncs).Parse(entityTemplate))
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to generate entity file: %w", err)
	}

	fmt.Printf("✅ Entity created: %s\n", filePath)
	fmt.Printf("📝 Entity: %s\n", entityName)
	fmt.Printf("🗂️  Table: %s\n", tableName)

	// Show entity features
	if len(fields) > 0 {
		fmt.Printf("📋 Entity Features:\n")

		// Check for associations
		hasAssociations := false
		for _, field := range fields {
			if field.IsForeignKey {
				hasAssociations = true
				refEntity := getStructName(field.FKReference)
				fmt.Printf("  - %s association (belongs to %s)\n", refEntity, refEntity)
			}
		}

		// Check for indexes
		hasIndexes := false
		for _, field := range fields {
			if field.HasIndex {
				hasIndexes = true
				fmt.Printf("  - Index on %s field\n", field.Name)
			}
		}

		if !hasAssociations && !hasIndexes {
			fmt.Printf("  - Basic CRUD entity with validation\n")
		}

		fmt.Printf("  - Soft deletes enabled\n")
		fmt.Printf("  - JSON serialization ready\n")
		fmt.Printf("  - Validation tags included\n")
	}

	return nil
}

func createSeeder(seederName, tableName, depsStr string) {
	if !strings.HasSuffix(seederName, "Seeder") {
		seederName += "Seeder"
	}

	fileName := fmt.Sprintf("%s.go", toSnakeCase(seederName))

	// Create seeders directory if not exists
	seedersDir := "internal/seeders"
	if err := os.MkdirAll(seedersDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create seeders directory: %v\n", err)
		os.Exit(1)
	}

	filePath := filepath.Join(seedersDir, fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("❌ Seeder file already exists: %s\n", filePath)
		os.Exit(1)
	}

	// Parse dependencies
	var dependencies []string
	if depsStr != "" {
		depsList := strings.Split(depsStr, ",")
		for _, dep := range depsList {
			dep = strings.TrimSpace(dep)
			if dep != "" {
				// Ensure dependency ends with "Seeder"
				if !strings.HasSuffix(dep, "Seeder") {
					dep += "Seeder"
				}
				dependencies = append(dependencies, dep)
			}
		}
	}

	data := SeederData{
		ClassName:    seederName,
		TableName:    tableName,
		Dependencies: dependencies,
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("❌ Failed to create seeder file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Execute template
	tmpl := template.Must(template.New("seeder").Parse(seederTemplate))
	if err := tmpl.Execute(file, data); err != nil {
		fmt.Printf("❌ Failed to generate seeder file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Seeder created: %s\n", filePath)
	fmt.Printf("📝 Class: %s\n", data.ClassName)
	if tableName != "" {
		fmt.Printf("🗂️  Table: %s\n", tableName)
	}
	if len(dependencies) > 0 {
		fmt.Printf("🔗 Dependencies: %s\n", strings.Join(dependencies, ", "))
	}
}

func createModel(modelName, table, fieldList string) {
	// Generate entity struct name
	entityName := toPascalCase(modelName)

	// Use TABLE parameter if provided, otherwise auto-generate
	var tableName string
	if table != "" {
		tableName = table // Use provided table name
		fmt.Printf("📋 Using specified table: %s\n", tableName)
	} else {
		tableName = strings.ToLower(toSnakeCase(entityName)) + "s" // Auto-generate: posts, users, etc.
		fmt.Printf("📋 Auto-generated table: %s\n", tableName)
	}

	fileName := fmt.Sprintf("%s.go", strings.ToLower(entityName))

	// Create entity directory if not exists
	entityDir := "internal/entity"
	if err := os.MkdirAll(entityDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create entity directory: %v\n", err)
		os.Exit(1)
	}

	filePath := filepath.Join(entityDir, fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("❌ Entity file already exists: %s\n", filePath)
		os.Exit(1)
	}

	// Use enhanced parseFields function (same as migration)
	parsedFields := parseFields(fieldList)

	// Create entity data
	data := EntityData{
		EntityName: entityName,
		TableName:  tableName, // Use specified or auto-generated table name
		Fields:     parsedFields,
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("❌ Failed to create entity file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Execute template
	tmpl := template.Must(template.New("entity").Funcs(templateFuncs).Parse(entityTemplate))
	if err := tmpl.Execute(file, data); err != nil {
		fmt.Printf("❌ Failed to generate entity file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Entity created: %s\n", filePath)
	fmt.Printf("📝 Entity: %s\n", entityName)
	fmt.Printf("🗂️  Table: %s\n", tableName)

	// Enhanced field summary (same as migration)
	if len(parsedFields) > 0 {
		fmt.Printf("📋 Fields:\n")
		for _, field := range parsedFields {
			extras := []string{}
			if field.HasIndex {
				extras = append(extras, "indexed")
			}
			if field.IsForeignKey {
				extras = append(extras, fmt.Sprintf("FK->%s", field.FKReference))
			}

			extraStr := ""
			if len(extras) > 0 {
				extraStr = fmt.Sprintf(" (%s)", strings.Join(extras, ", "))
			}

			fmt.Printf("  - %s: %s%s\n", field.Name, field.Type, extraStr)
		}

		// Show entity features (same as autoCreateEntity)
		fmt.Printf("📋 Entity Features:\n")

		// Check for associations
		hasAssociations := false
		for _, field := range parsedFields {
			if field.IsForeignKey {
				hasAssociations = true
				refEntity := getStructName(field.FKReference)
				fmt.Printf("  - %s association (belongs to %s)\n", refEntity, refEntity)
			}
		}

		// Check for indexes
		hasIndexes := false
		for _, field := range parsedFields {
			if field.HasIndex {
				hasIndexes = true
				fmt.Printf("  - Index on %s field\n", field.Name)
			}
		}

		if !hasAssociations && !hasIndexes {
			fmt.Printf("  - Basic CRUD entity with validation\n")
		}

		fmt.Printf("  - Soft deletes enabled\n")
		fmt.Printf("  - JSON serialization ready\n")
		fmt.Printf("  - Validation tags included\n")
	}
}
func createPackage(packageName string) {
	// Convert to lowercase for package name
	pkgName := toSnakeCase(packageName)
	entityName := toPascalCase(packageName)

	// Create package directory
	packageDir := filepath.Join("internal", pkgName)
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create package directory: %v\n", err)
		os.Exit(1)
	}

	// Check if package already exists
	files := []string{"handler.go", "port.go", "repository.go", "usecase.go"}
	for _, file := range files {
		if _, err := os.Stat(filepath.Join(packageDir, file)); err == nil {
			fmt.Printf("❌ Package '%s' already exists (found %s)\n", pkgName, file)
			os.Exit(1)
		}
	}

	packageData := PackageData{
		PackageName: pkgName,
		EntityName:  entityName,
	}

	// Create handler.go
	if err := createFileFromTemplate(
		filepath.Join(packageDir, "handler.go"),
		handlerTemplate,
		packageData,
	); err != nil {
		fmt.Printf("❌ Failed to create handler.go: %v\n", err)
		os.Exit(1)
	}

	// Create port.go
	if err := createFileFromTemplate(
		filepath.Join(packageDir, "port.go"),
		portTemplate,
		packageData,
	); err != nil {
		fmt.Printf("❌ Failed to create port.go: %v\n", err)
		os.Exit(1)
	}

	// Create repository.go
	if err := createFileFromTemplate(
		filepath.Join(packageDir, "repository.go"),
		repositoryTemplate,
		packageData,
	); err != nil {
		fmt.Printf("❌ Failed to create repository.go: %v\n", err)
		os.Exit(1)
	}

	// Create usecase.go
	if err := createFileFromTemplate(
		filepath.Join(packageDir, "usecase.go"),
		usecaseTemplate,
		packageData,
	); err != nil {
		fmt.Printf("❌ Failed to create usecase.go: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Package created: internal/%s/\n", pkgName)
	fmt.Printf("📁 Files created:\n")
	fmt.Printf("  - internal/%s/handler.go\n", pkgName)
	fmt.Printf("  - internal/%s/port.go\n", pkgName)
	fmt.Printf("  - internal/%s/repository.go\n", pkgName)
	fmt.Printf("  - internal/%s/usecase.go\n", pkgName)
	fmt.Printf("🎯 Entity: %s\n", entityName)
}

func createFileFromTemplate(filePath, templateContent string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("template").Funcs(templateFuncs).Parse(templateContent))
	return tmpl.Execute(file, data)
}

func runMigrations() {
	fmt.Println("⬆️  Running migrations...")

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		fmt.Printf("❌ Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Initialize database using factory
	factory := pkgDatabase.NewDatabaseFactory()
	dbConfig := cfg.GetDatabaseConfig()

	db, err := factory.CreateDatabase(dbConfig)
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s database: %v\n", cfg.Database.Type, err)
		os.Exit(1)
	}

	fmt.Printf("📊 Using %s database\n", cfg.Database.Type)

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		fmt.Printf("❌ Migration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Migrations completed successfully")
}

func rollbackMigrations(count string) {
	fmt.Printf("⬇️  Rolling back %s migration(s)...\n", count)

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		fmt.Printf("❌ Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Initialize database using factory
	factory := pkgDatabase.NewDatabaseFactory()
	dbConfig := cfg.GetDatabaseConfig()

	db, err := factory.CreateDatabase(dbConfig)
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s database: %v\n", cfg.Database.Type, err)
		os.Exit(1)
	}

	fmt.Printf("📊 Using %s database\n", cfg.Database.Type)

	// Rollback migrations
	if err := db.RollbackMigrations(count); err != nil {
		fmt.Printf("❌ Rollback failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Rollback completed successfully")
}

func showMigrationStatus() {
	fmt.Println("📊 Checking migration status...")

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		fmt.Printf("❌ Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Initialize database using factory
	factory := pkgDatabase.NewDatabaseFactory()
	dbConfig := cfg.GetDatabaseConfig()

	db, err := factory.CreateDatabase(dbConfig)
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s database: %v\n", cfg.Database.Type, err)
		os.Exit(1)
	}

	fmt.Printf("📊 Using %s database\n", cfg.Database.Type)

	// Show migration status
	if err := db.GetMigrationStatus(); err != nil {
		fmt.Printf("❌ Failed to get migration status: %v\n", err)
		os.Exit(1)
	}
}

func runSeeders(seederName string) {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		fmt.Printf("❌ Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Initialize database using factory
	factory := pkgDatabase.NewDatabaseFactory()
	dbConfig := cfg.GetDatabaseConfig()

	db, err := factory.CreateDatabase(dbConfig)
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s database: %v\n", cfg.Database.Type, err)
		os.Exit(1)
	}

	if seederName == "list" {
		fmt.Println("📋 Listing seeders...")
		fmt.Printf("📊 Using %s database\n", cfg.Database.Type)

		if err := db.ListSeeders(); err != nil {
			fmt.Printf("❌ Failed to list seeders: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println("🌱 Running seeders...")
	fmt.Printf("📊 Using %s database\n", cfg.Database.Type)

	// Run seeders
	if err := db.SeedData(seederName); err != nil {
		fmt.Printf("❌ Seeding failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Seeding completed successfully")
}

func showHelp() {
	fmt.Println("🎨 Go Clean Gin - Artisan CLI (Laravel Style)")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/artisan/main.go -action=<action> [options]")
	fmt.Println("")
	fmt.Println("Available Actions:")
	fmt.Println("  make:migration     Create a new migration file")
	fmt.Println("  make:seeder        Create a new seeder file")
	fmt.Println("  make:model         Create a new entity model file")
	fmt.Println("  make:package       Create a new package with handler, usecase, repository, port")
	fmt.Println("  migrate            Run pending migrations")
	fmt.Println("  migrate:rollback   Rollback migrations")
	fmt.Println("  migrate:status     Show migration status")
	fmt.Println("  db:seed            Run database seeders")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -name string       Migration/Seeder/Model/Package name")
	fmt.Println("  -table string      Table name")
	fmt.Println("  -create            Create table migration")
	fmt.Println("  -fields string     Fields (name:string,email:string)")
	fmt.Println("  -count int         Number of migrations to rollback (default: 1)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Create table migration")
	fmt.Println("  go run cmd/artisan/main.go -action=make:migration -name=create_users_table -create -table=users -fields=\"name:string,email:string\"")
	fmt.Println("")
	fmt.Println("  # Create entity model")
	fmt.Println("  go run cmd/artisan/main.go -action=make:model -name=User -fields=\"name:string,email:string,age:int\"")
	fmt.Println("")
	fmt.Println("  # Create package (handler, usecase, repository, port)")
	fmt.Println("  go run cmd/artisan/main.go -action=make:package -name=Product")
	fmt.Println("")
	fmt.Println("  # Add column migration")
	fmt.Println("  go run cmd/artisan/main.go -action=make:migration -name=add_phone_to_users -table=users -fields=\"phone:string\"")
	fmt.Println("")
	fmt.Println("  # Run migrations")
	fmt.Println("  go run cmd/artisan/main.go -action=migrate")
	fmt.Println("")
	fmt.Println("  # Rollback last 2 migrations")
	fmt.Println("  go run cmd/artisan/main.go -action=migrate:rollback -count=2")
	fmt.Println("")
	fmt.Println("  # Create seeder")
	fmt.Println("  go run cmd/artisan/main.go -action=make:seeder -name=UserSeeder -table=users")
	fmt.Println("")
	fmt.Println("  # Create seeder with dependencies")
	fmt.Println("  go run cmd/artisan/main.go -action=make:seeder -name=ProductSeeder -table=products -deps=\"UserSeeder\"")
	fmt.Println("  go run cmd/artisan/main.go -action=make:seeder -name=OrderSeeder -table=orders -deps=\"UserSeeder,ProductSeeder\"")
	fmt.Println("")
	fmt.Println("  # List all seeders")
	fmt.Println("  go run cmd/artisan/main.go -action=db:seed -name=list")
}

// Helper types and functions
type MigrationData struct {
	ClassName   string
	TableName   string
	Timestamp   string
	Description string
	Fields      []Field
	Version     string
}

type Field struct {
	ClassName    string
	Name         string
	Type         string
	HasIndex     bool
	IsForeignKey bool
	FKReference  string // table name that reference
}

type SeederData struct {
	ClassName    string
	TableName    string
	Dependencies []string // add this field
}

type EntityData struct {
	EntityName string
	TableName  string
	Fields     []Field
}

type PackageData struct {
	PackageName string
	EntityName  string
}

func parseFields(fieldList string) []Field {
	var parsedFields []Field
	if fieldList == "" {
		return parsedFields
	}

	fieldPairs := strings.Split(fieldList, ",")

	for _, pair := range fieldPairs {
		// split field_name:type|options - use SplitN to split only the first ":"
		mainParts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(mainParts) < 2 {
			continue
		}

		fieldName := strings.TrimSpace(mainParts[0])
		typeAndOptions := strings.TrimSpace(mainParts[1])

		// split type and options (type|index or type|fk:table)
		typeParts := strings.Split(typeAndOptions, "|")
		fieldType := strings.TrimSpace(typeParts[0])

		field := Field{
			Name:         fieldName,
			Type:         fieldType,
			HasIndex:     false,
			IsForeignKey: false,
			FKReference:  "",
		}

		// check options
		if len(typeParts) > 1 {
			for i := 1; i < len(typeParts); i++ {
				option := strings.TrimSpace(typeParts[i])

				if option == "index" {
					field.HasIndex = true
				} else if strings.HasPrefix(option, "fk:") {
					field.IsForeignKey = true
					field.FKReference = strings.TrimPrefix(option, "fk:")
				}
			}
		}

		parsedFields = append(parsedFields, field)
	}

	return parsedFields
}

// Template functions
var templateFuncs = template.FuncMap{
	"toGoType":         toGoType,
	"toPascalCase":     toPascalCase,
	"toCamelCase":      toCamelCase,
	"getGormTag":       getGormTag,
	"getValidationTag": getValidationTag,
	"hasDecimalField":  hasDecimalField,
	"getStructName":    getStructName,
	"hasIndexField":    hasIndexField,
	"hasFKField":       hasFKField,
	"toLowerFirst":     toLowerFirst,
}

func toPascalCase(s string) string {
	// First, split by common separators
	words := strings.FieldsFunc(s, func(c rune) bool {
		return c == '_' || c == '-' || c == ' '
	})

	// If we have a single word and it's camelCase, split it by capital letters
	if len(words) == 1 && words[0] == s {
		// Split camelCase into words
		var camelWords []string
		var current strings.Builder

		for i, r := range s {
			if i > 0 && 'A' <= r && r <= 'Z' {
				if current.Len() > 0 {
					camelWords = append(camelWords, current.String())
					current.Reset()
				}
			}
			current.WriteRune(r)
		}
		if current.Len() > 0 {
			camelWords = append(camelWords, current.String())
		}

		words = camelWords
	}

	caser := cases.Title(language.English)
	for i, word := range words {
		words[i] = caser.String(strings.ToLower(word))
	}

	return strings.Join(words, "")
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func toCamelCase(s string) string {
	// First convert to PascalCase
	pascalCase := toPascalCase(s)
	// Then convert first letter to lowercase
	if len(pascalCase) > 0 {
		return strings.ToLower(pascalCase[:1]) + pascalCase[1:]
	}
	return pascalCase
}

func toGoType(fieldType string) string {
	switch strings.ToLower(fieldType) {
	case "string":
		return "string"
	case "text":
		return "string"
	case "int", "integer":
		return "int"
	case "int64", "bigint":
		return "int64"
	case "float", "float64":
		return "float64"
	case "decimal":
		return "decimal.Decimal"
	case "bool", "boolean":
		return "bool"
	case "uuid":
		return "uuid.UUID"
	case "timestamp", "time", "date", "datetime":
		return "time.Time"
	case "json", "jsonb":
		return "map[string]interface{}"
	default:
		return "string"
	}
}

func getGormTag(field Field) string {
	tags := []string{}

	// Basic type tags
	switch strings.ToLower(field.Type) {
	case "string":
		tags = append(tags, "type:varchar(255)", "not null")
	case "text":
		tags = append(tags, "type:text")
	case "int", "integer":
		tags = append(tags, "type:integer", "not null")
	case "int64", "bigint":
		tags = append(tags, "type:bigint", "not null")
	case "float":
		tags = append(tags, "type:float", "not null")
	case "float64":
		tags = append(tags, "type:double", "not null")
	case "decimal":
		tags = append(tags, "type:decimal(10,2)", "not null")
		// case "bool", "boolean":
		// 	tags = append(tags, "default:false")
		// case "uuid":
		// 	tags = append(tags, "type:varchar(36)", "not null")
		// case "time":
		// 	tags = append(tags, "type:time")
		// case "date":
		// 	tags = append(tags, "type:date")
	case "json":
		tags = append(tags, "type:json", "default:'{}'")
	case "jsonb":
		tags = append(tags, "type:jsonb", "default:'{}'")
	default:
		tags = append(tags, "not null")
	}

	// Add index tag
	if field.HasIndex || field.IsForeignKey {
		tags = append(tags, "index")
	}

	// Add foreign key constraint
	if field.IsForeignKey {
		tags = append(tags, "constraint:OnUpdate:CASCADE,OnDelete:SET NULL")
	}

	return strings.Join(tags, ";")
}

func getValidationTag(fieldType string) string {
	switch strings.ToLower(fieldType) {
	case "string":
		return "required,min=1,max=255"
	case "text":
		return "required"
	case "int", "integer":
		return "required,min=0"
	case "int64", "bigint":
		return "required,min=0"
	case "float", "float64":
		return "required,min=0"
	case "decimal":
		return "required,min=0"
	case "uuid":
		return "required"
	default:
		return "required"
	}
}

func hasDecimalField(fields []Field) bool {
	for _, field := range fields {
		if strings.ToLower(field.Type) == "decimal" {
			return true
		}
	}
	return false
}

func getStructName(tableName string) string {
	// if table name start with tb_ then remove it
	tableName = strings.TrimPrefix(tableName, "tb_")

	tableName = singularize(tableName)

	// convert to PascalCase
	return toPascalCase(tableName)
}

// singularize convert plural to singular (simple format)
func singularize(word string) string {
	// basic rules for English pluralization
	if strings.HasSuffix(word, "ies") {
		// categories -> category, companies -> company
		return strings.TrimSuffix(word, "ies") + "y"
	}
	if strings.HasSuffix(word, "es") && len(word) > 2 {
		// boxes -> box, dishes -> dish
		return strings.TrimSuffix(word, "es")
	}
	if strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss") {
		// users -> user, products -> product (but not address -> addres)
		return strings.TrimSuffix(word, "s")
	}

	// if not match any rule then return original word
	return word
}

func hasIndexField(fields []Field) bool {
	for _, field := range fields {
		if field.HasIndex {
			return true
		}
	}
	return false
}

func hasFKField(fields []Field) bool {
	for _, field := range fields {
		if field.IsForeignKey {
			return true
		}
	}
	return false
}

func toLowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// Templates
const migrationTemplate = `package migrations

import (
	"gorm.io/gorm"
)

// {{.ClassName}} migration
type {{.ClassName}} struct{}

// Up runs the migration
func (m *{{.ClassName}}) Up(db *gorm.DB) error {
	// TODO: Implement your migration logic here
	return nil
}

// Down rolls back the migration  
func (m *{{.ClassName}}) Down(db *gorm.DB) error {
	// TODO: Implement your rollback logic here
	return nil
}

// Description returns migration description
func (m *{{.ClassName}}) Description() string {
	return "{{.Description}}"
}

// Version returns migration version
func (m *{{.ClassName}}) Version() string {
	return "{{.Version}}"
}

// Auto-register migration
func init() {
	Register(&{{.ClassName}}{})
}
`

const createTableTemplate = `package migrations

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	{{- if hasDecimalField .Fields}}
	"github.com/shopspring/decimal"
	{{- end}}
)

// {{getStructName .TableName}} entity struct for migration
type {{getStructName .TableName}} struct {
	ID        int       ` + "`gorm:\"primaryKey\"`" + `
	UUID      uuid.UUID  ` + "`gorm:\"type:varchar(36);unique;index;not null\"`" + `
	{{- range .Fields}}
	{{toPascalCase .Name}} {{toGoType .Type}} ` + "`gorm:\"{{getGormTag .}}\"`" + `
	{{- end}}
	{{- range .Fields}}
	{{- if .IsForeignKey}}
	{{getStructName .FKReference}} {{getStructName .FKReference}} ` + "`json:\"{{getStructName .FKReference | toLowerFirst}},omitempty\" gorm:\"foreignKey:{{toPascalCase .Name}};references:ID\"`" + `
	{{- end}}
	{{- end}}
	CreatedAt time.Time      ` + "`gorm:\"autoCreateTime;not null;default:CURRENT_TIMESTAMP(3)\"`" + `
	UpdatedAt time.Time      ` + "`gorm:\"autoUpdateTime;not null;default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)\"`" + `
	DeletedAt gorm.DeletedAt ` + "`gorm:\"index\"`" + `
}

// TableName returns the table name for GORM
func ({{getStructName .TableName}}) TableName() string {
	return "{{.TableName}}"
}

// {{.ClassName}} migration - Create {{.TableName}} table
type {{.ClassName}} struct{}

// Up creates the {{.TableName}} table using the {{getStructName .TableName}} struct
func (m *{{.ClassName}}) Up(db *gorm.DB) error {
	return db.AutoMigrate(&{{getStructName .TableName}}{})
}

// Down drops the {{.TableName}} table
func (m *{{.ClassName}}) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&{{getStructName .TableName}}{})
}

// Description returns migration description
func (m *{{.ClassName}}) Description() string {
	return "Create {{.TableName}} table"
}

// Version returns migration version
func (m *{{.ClassName}}) Version() string {
	return "{{.Version}}"
}

// Auto-register migration
func init() {
	Register(&{{.ClassName}}{})
}
`

const alterTableTemplate = `package migrations

import (
	"gorm.io/gorm"
	{{- if hasDecimalField .Fields}}
	"github.com/shopspring/decimal"
	{{- end}}
)

// {{.ClassName}} migration - Modify {{.TableName}} table
type {{.ClassName}} struct{}

{{- range .Fields}}
// {{.ClassName}}{{toPascalCase .Name}} represents the new column structure
type {{$.ClassName}}{{toPascalCase .Name}} struct {
	{{toPascalCase .Name}} {{toGoType .Type}} ` + "`gorm:\"{{getGormTag .}}\"`" + `
}

func ({{$.ClassName}}{{toPascalCase .Name}}) TableName() string {
	return "{{$.TableName}}"
}
{{- end}}

// Up adds columns to the {{.TableName}} table
func (m *{{.ClassName}}) Up(db *gorm.DB) error {
	{{- range .Fields}}
	// Add {{.Name}} column
	if err := db.Migrator().AddColumn(&{{$.ClassName}}{{toPascalCase .Name}}{}, "{{.Name}}"); err != nil {
		return err
	}
	{{- end}}
	
	return nil
}

// Down removes columns from the {{.TableName}} table
func (m *{{.ClassName}}) Down(db *gorm.DB) error {
	{{- range .Fields}}
	// Drop {{.Name}} column
	if err := db.Migrator().DropColumn(&{{$.ClassName}}{{toPascalCase .Name}}{}, "{{.Name}}"); err != nil {
		return err
	}
	{{- end}}
	
	return nil
}

// Description returns migration description
func (m *{{.ClassName}}) Description() string {
	return "{{.Description}}"
}

// Version returns migration version
func (m *{{.ClassName}}) Version() string {
	return "{{.Version}}"
}

// Auto-register migration
func init() {
	Register(&{{.ClassName}}{})
}
`

const seederTemplate = `package seeders

import (
	"gorm.io/gorm"
	"go-starter/internal/entity"
	"go-starter/pkg/logger"
	"go.uber.org/zap"
)

// {{.ClassName}} seeds the {{.TableName}} table
type {{.ClassName}} struct{}

// Run executes the seeder
func (s *{{.ClassName}}) Run(db *gorm.DB) error {
	logger.Info("Running {{.ClassName}}...")

	// Check if data already exists
	{{- if .TableName}}
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM {{.TableName}}").Scan(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logger.Info("{{.TableName}} already exist, skipping {{.ClassName}}")
		return nil
	}
	{{- end}}

	// TODO: Implement your seeding logic here
	// Example:
	{{- if .Dependencies}}
	//
	// This seeder depends on: {{range $i, $dep := .Dependencies}}{{if $i}}, {{end}}{{$dep}}{{end}}
	// You can safely reference data created by those seeders
	//
	{{- end}}
	// data := []entity.Model{
	//     {Field1: "value1", Field2: "value2"},
	//     {Field1: "value3", Field2: "value4"},
	// }
	//
	// return db.Create(&data).Error

	logger.Info("{{.ClassName}} completed successfully")
	return nil
}

// Name returns seeder name
func (s *{{.ClassName}}) Name() string {
	return "{{.ClassName}}"
}

// Dependencies returns list of seeders that must run before this seeder
func (s *{{.ClassName}}) Dependencies() []string {
	{{- if .Dependencies}}
	return []string{
		{{- range .Dependencies}}
		"{{.}}",
		{{- end}}
	}
	{{- else}}
	return []string{} // No dependencies
	{{- end}}
}

// Auto-register seeder
func init() {
	Register(&{{.ClassName}}{})
}
`

// Fix entityTemplate - add association fields like createTableTemplate
const entityTemplate = `package entity

import (
	"time"

	"github.com/google/uuid"
	{{- if hasDecimalField .Fields}}
	"github.com/shopspring/decimal"
	{{- end}}
	"gorm.io/gorm"
)

// {{.EntityName}} represents a {{.EntityName}} entity
type {{.EntityName}} struct {
	ID        int       ` + "`json:\"-\" gorm:\"primaryKey\"`" + `
	UUID      uuid.UUID  ` + "`json:\"uuid\" gorm:\"type:varchar(36);unique;index;not null\"`" + `
	{{- range .Fields}}
	{{toPascalCase .Name}} {{toGoType .Type}} ` + "`json:\"{{.Name}}\" gorm:\"{{getGormTag .}}\"`" + `
	{{- end}}
	{{- range .Fields}}
	{{- if .IsForeignKey}}
	{{getStructName .FKReference}} {{getStructName .FKReference}} ` + "`json:\"{{getStructName .FKReference | toLowerFirst}},omitempty\" gorm:\"foreignKey:{{toPascalCase .Name}};references:ID\"`" + `
	{{- end}}
	{{- end}}
	CreatedAt time.Time      ` + "`json:\"created_at\" gorm:\"autoCreateTime;default:CURRENT_TIMESTAMP(3);not null\"`" + `
	UpdatedAt time.Time      ` + "`json:\"updated_at\" gorm:\"autoUpdateTime;default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);not null\"`" + `
	DeletedAt gorm.DeletedAt ` + "`json:\"-\" gorm:\"index\"`" + `
}

// TableName returns the table name for GORM
func ({{.EntityName}}) TableName() string {
	return "{{.TableName}}"
}

// BeforeCreate is a hook that runs before creating a {{.EntityName}}
func (e *{{.EntityName}}) BeforeCreate(tx *gorm.DB) (err error) {
	e.UUID = uuid.New()
	return
}

// Create{{.EntityName}}Request represents a request to create a {{.EntityName}}
type Create{{.EntityName}}Request struct {
	{{- range .Fields}}
	{{toPascalCase .Name}} {{toGoType .Type}} ` + "`json:\"{{.Name}}\" validate:\"{{getValidationTag .Type}}\"`" + `
	{{- end}}
}

// Update{{.EntityName}}Request represents a request to update a {{.EntityName}}
type Update{{.EntityName}}Request struct {
	{{- range .Fields}}
	{{toPascalCase .Name}} *{{toGoType .Type}} ` + "`json:\"{{.Name}},omitempty\" validate:\"omitempty,{{getValidationTag .Type}}\"`" + `
	{{- end}}
}

// {{.EntityName}}Filter represents filters for {{.EntityName}} queries
type {{.EntityName}}Filter struct {
	{{- range .Fields}}
	{{- if eq .Type "string"}}
	{{toPascalCase .Name}} string ` + "`form:\"{{.Name}}\"`" + `
	{{- end}}
	{{- end}}
	Search string ` + "`form:\"search\"`" + `
	Page   int    ` + "`form:\"page\" validate:\"min=1\"`" + `
	Limit  int    ` + "`form:\"limit\" validate:\"min=1,max=100\"`" + `
}

`

// Package templates - Simple structure without CRUD
const handlerTemplate = `package {{.PackageName}}

import (
	"go-starter/pkg/errors"
	"go-starter/pkg/logger"
	"go-starter/pkg/response"
	"go-starter/pkg/validator"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type {{.EntityName}}Handler struct {
	usecase {{.EntityName}}Usecase
}

func New{{.EntityName}}Handler(usecase {{.EntityName}}Usecase) *{{.EntityName}}Handler {
	return &{{.EntityName}}Handler{
		usecase: usecase,
	}
}

// TODO: Add your handler methods here
// Example:
// func (h *{{.EntityName}}Handler) SomeMethod(c *gin.Context) {
//     // Implementation here
//     // h.usecase.SomeMethod(ctx)
// }
`

const portTemplate = `package {{.PackageName}}

import (
	"context"
)

// {{.EntityName}}Usecase defines the business logic interface for {{.PackageName}}
type {{.EntityName}}Usecase interface {
	// TODO: Add your usecase methods here
	// Example:
	// SomeMethod(ctx context.Context) error
}

// {{.EntityName}}Repository defines the data access interface for {{.PackageName}}
type {{.EntityName}}Repository interface {
	// TODO: Add your repository methods here
	// Example:
	// SomeMethod(ctx context.Context) error
}
`

const repositoryTemplate = `package {{.PackageName}}

import (
	"context"

	"gorm.io/gorm"
)

type {{toCamelCase .EntityName}}Repository struct {
	db *gorm.DB
}

func New{{.EntityName}}Repository(db *gorm.DB) {{.EntityName}}Repository {
	return &{{toCamelCase .EntityName}}Repository{
		db: db,
	}
}

// TODO: Add your repository methods here
// Example:
// func (r *{{toCamelCase .EntityName}}Repository) SomeMethod(ctx context.Context) error {
//     return r.db.WithContext(ctx).Error
// }
`

const usecaseTemplate = `package {{.PackageName}}

import (
	"context"
	"go-starter/pkg/errors"
	"go-starter/pkg/logger"

	"go.uber.org/zap"
)

type {{toCamelCase .EntityName}}Usecase struct {
	repo {{.EntityName}}Repository
}

func New{{.EntityName}}Usecase(repo {{.EntityName}}Repository) {{.EntityName}}Usecase {
	return &{{toCamelCase .EntityName}}Usecase{
		repo: repo,
	}
}

// TODO: Add your usecase methods here
// Example:
// func (u *{{toCamelCase .EntityName}}Usecase) SomeMethod(ctx context.Context) error {
//     logger.Info("Executing SomeMethod for {{.EntityName}}")
//     
//     if err := u.repo.SomeMethod(ctx); err != nil {
//         logger.Error("Failed to execute SomeMethod", zap.Error(err))
//         return errors.Wrap(err, errors.ErrInternal, "Failed to execute SomeMethod", 500)
//     }
//     
//     return nil
// }
`
