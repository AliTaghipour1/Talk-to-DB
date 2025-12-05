package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MySqlConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type DatabaseMySqlImpl struct {
	db     *sql.DB
	config MySqlConfig
}

func (d *DatabaseMySqlImpl) connect() error {
	// Build MySQL connection string
	// Format: username:password@tcp(host:port)/database?parseTime=true
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		d.config.User,
		d.config.Password,
		d.config.Host,
		d.config.Port,
		d.config.Database,
	)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Verify connection by pinging
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	// Set connection pool settings for thread safety
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	d.db = db
	return nil
}

func (d *DatabaseMySqlImpl) GetTables() ([]Table, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	// Query to get all table names from the current database
	query := `
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = ? 
		AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := d.db.Query(query, d.config.Database)
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
		columns, err := d.getColumns(tableName)
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

func (d *DatabaseMySqlImpl) getColumns(tableName string) ([]Column, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := d.db.Query(query, d.config.Database, tableName)
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

func (d *DatabaseMySqlImpl) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if d.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	// Use parameterized queries to prevent SQL injection
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return rows, nil
}

// Close closes the database connection
func (d *DatabaseMySqlImpl) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func NewDatabaseMySqlImpl(config MySqlConfig) (Database, error) {
	result := &DatabaseMySqlImpl{
		config: config,
	}

	err := result.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL database instance: %w", err)
	}

	return result, nil
}
