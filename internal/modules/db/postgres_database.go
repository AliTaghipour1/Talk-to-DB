package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	Schema   string // Default to "public" if empty
}

type DatabasePostgresImpl struct {
	db     *sql.DB
	config PostgresConfig
}

func (d *DatabasePostgresImpl) connect() error {
	// Set default schema if not provided
	schema := d.config.Schema
	if schema == "" {
		schema = "public"
	}

	// Set default SSL mode if not provided
	sslMode := d.config.SSLMode
	if sslMode == "" {
		sslMode = "disable" // Default to disable for local development
	}

	// Build PostgreSQL connection string
	// Format: host=... port=... user=... password=... dbname=... sslmode=...
	connParts := []string{
		fmt.Sprintf("host=%s", d.config.Host),
		fmt.Sprintf("port=%s", d.config.Port),
		fmt.Sprintf("user=%s", d.config.User),
		fmt.Sprintf("password=%s", d.config.Password),
		fmt.Sprintf("dbname=%s", d.config.DBName),
		fmt.Sprintf("sslmode=%s", sslMode),
	}
	connStr := strings.Join(connParts, " ")

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Verify connection by pinging
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	// Set connection pool settings for thread safety
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	d.db = db
	return nil
}

func (d *DatabasePostgresImpl) GetTables() (Tables, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	// Determine schema to use
	schema := d.config.Schema
	if schema == "" {
		schema = "public"
	}

	// Query to get all table names from the specified schema
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = $1 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := d.db.Query(query, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}

		// Get columns for this table
		columns, err := d.getColumns(tableName, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}

		tables = append(tables, Table{
			Name:    tableName,
			Columns: columns,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	return tables, nil
}

func (d *DatabasePostgresImpl) getColumns(tableName, schema string) ([]Column, error) {
	query := `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`

	rows, err := d.db.Query(query, schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(&col.Name, &col.DataType); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column rows: %w", err)
	}

	return columns, nil
}

func (d *DatabasePostgresImpl) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	// Use parameterized queries ($1, $2, ...) to prevent SQL injection
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return rows, nil
}

// Close closes the database connection
func (d *DatabasePostgresImpl) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetSchemas returns all available schemas in the database
func (d *DatabasePostgresImpl) GetSchemas() ([]string, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
		ORDER BY schema_name
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, fmt.Errorf("failed to scan schema name: %w", err)
		}
		schemas = append(schemas, schemaName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schema rows: %w", err)
	}

	return schemas, nil
}

func NewDatabasePostgresImpl(config PostgresConfig) (Database, error) {
	result := &DatabasePostgresImpl{
		config: config,
	}

	err := result.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL database instance: %w", err)
	}

	return result, nil
}
