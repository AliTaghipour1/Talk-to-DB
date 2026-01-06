package repo

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type Table struct {
	ID          int
	Name        string
	Description string
	Columns     []Column
}

type Column struct {
	ID          int
	Name        string
	DataType    string
	Description string
}

type Database struct {
	ID          int
	Name        string
	Description string
	Tables      []Table
}

func (s Database) Scheme() string {
	indent, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(indent)
}

type fieldType int8

const (
	ColumnFieldType   = fieldType(1)
	TableFieldType    = fieldType(2)
	DatabaseFieldType = fieldType(3)
)

type DatabaseRepo interface {
	CreateNewDatabase(database *Database) (int, error)
	GetDatabase(ID int) (Database, error)
	GetAllDatabases() ([]Database, error)
	SetDescription(dbID int, desc string, fieldID int, fieldType fieldType) error
}

// persistenceData represents the structure saved to JSON file
type persistenceData struct {
	Databases      map[int]*Database `json:"databases"`
	NextDatabaseID int               `json:"next_database_id"`
	NextTableID    int               `json:"next_table_id"`
	NextColumnID   int               `json:"next_column_id"`
}

type DatabaseRepoMapImpl struct {
	databaseMap    map[int]*Database
	nextDatabaseID int
	nextTableID    int
	nextColumnID   int
	filePath       string
	mu             sync.RWMutex
}

// NewDatabaseRepoMapImpl creates a new repository instance with file persistence
func NewDatabaseRepoMapImpl(filePath string) DatabaseRepo {
	repo := &DatabaseRepoMapImpl{
		databaseMap:    make(map[int]*Database),
		nextDatabaseID: 1,
		nextTableID:    1,
		nextColumnID:   1,
		filePath:       filePath,
	}

	// Try to load existing data from file
	if err := repo.loadFromFile(); err != nil {
		// If file doesn't exist, start with empty data (this is OK)
		if !os.IsNotExist(err) {
			log.Println("Error loading database:", err)
		}
	}

	return repo
}

// loadFromFile loads repository data from JSON file
func (r *DatabaseRepoMapImpl) loadFromFile() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return err
	}

	var persistData persistenceData
	if err := json.Unmarshal(data, &persistData); err != nil {
		return fmt.Errorf("failed to unmarshal repository data: %w", err)
	}

	r.databaseMap = persistData.Databases
	r.nextDatabaseID = persistData.NextDatabaseID
	r.nextTableID = persistData.NextTableID
	r.nextColumnID = persistData.NextColumnID

	// Initialize maps if they're nil (for backward compatibility)
	if r.databaseMap == nil {
		r.databaseMap = make(map[int]*Database)
	}

	return nil
}

// saveToFile saves repository data to JSON file
// Note: Caller must hold the write lock (mu.Lock())
func (r *DatabaseRepoMapImpl) saveToFile() error {
	persistData := persistenceData{
		Databases:      r.databaseMap,
		NextDatabaseID: r.nextDatabaseID,
		NextTableID:    r.nextTableID,
		NextColumnID:   r.nextColumnID,
	}

	data, err := json.MarshalIndent(persistData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal repository data: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write repository data to file: %w", err)
	}

	return nil
}

// assignIDs assigns unique IDs to all entities in the database
func (r *DatabaseRepoMapImpl) assignIDs(db *Database) {
	// Assign database ID
	if db.ID == 0 {
		db.ID = r.nextDatabaseID
		r.nextDatabaseID++
	}

	// Assign table IDs
	for i := range db.Tables {
		if db.Tables[i].ID == 0 {
			db.Tables[i].ID = r.nextTableID
			r.nextTableID++
		}

		// Assign column IDs
		for j := range db.Tables[i].Columns {
			if db.Tables[i].Columns[j].ID == 0 {
				db.Tables[i].Columns[j].ID = r.nextColumnID
				r.nextColumnID++
			}
		}
	}
}

// validateUniqueIDs ensures all IDs are unique within their types (globally across all databases)
func (r *DatabaseRepoMapImpl) validateUniqueIDs(db *Database) error {
	// Check database ID uniqueness
	if _, exists := r.databaseMap[db.ID]; exists && db.ID != 0 {
		return fmt.Errorf("database with ID %d already exists", db.ID)
	}

	// First, check for duplicates within the new database itself
	newTableIDs := make(map[int]bool)
	for i, table := range db.Tables {
		if table.ID != 0 {
			if newTableIDs[table.ID] {
				return fmt.Errorf("duplicate table ID %d within the database", table.ID)
			}
			newTableIDs[table.ID] = true

			// Check for duplicate column IDs within the table
			newColumnIDs := make(map[int]bool)
			for _, column := range db.Tables[i].Columns {
				if column.ID != 0 {
					if newColumnIDs[column.ID] {
						return fmt.Errorf("duplicate column ID %d within table %s", column.ID, table.Name)
					}
					newColumnIDs[column.ID] = true
				}
			}
		}
	}

	// Collect all existing table IDs from all databases
	existingTableIDs := make(map[int]bool)
	for _, existingDB := range r.databaseMap {
		for _, table := range existingDB.Tables {
			if table.ID != 0 {
				existingTableIDs[table.ID] = true
			}
		}
	}

	// Collect all existing column IDs from all databases
	existingColumnIDs := make(map[int]bool)
	for _, existingDB := range r.databaseMap {
		for _, table := range existingDB.Tables {
			for _, column := range table.Columns {
				if column.ID != 0 {
					existingColumnIDs[column.ID] = true
				}
			}
		}
	}

	// Check new database's IDs against existing databases
	for _, table := range db.Tables {
		if table.ID != 0 {
			if existingTableIDs[table.ID] {
				return fmt.Errorf("table with ID %d already exists in another database", table.ID)
			}
		}
		for _, column := range table.Columns {
			if column.ID != 0 {
				if existingColumnIDs[column.ID] {
					return fmt.Errorf("column with ID %d already exists in another database", column.ID)
				}
			}
		}
	}

	return nil
}

func (r *DatabaseRepoMapImpl) CreateNewDatabase(database *Database) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a deep copy to avoid modifying the original
	dbCopy := *database
	dbCopy.Tables = make([]Table, len(database.Tables))
	for i, table := range database.Tables {
		dbCopy.Tables[i] = table
		dbCopy.Tables[i].Columns = make([]Column, len(table.Columns))
		copy(dbCopy.Tables[i].Columns, table.Columns)
	}

	// Assign IDs if not provided
	r.assignIDs(&dbCopy)

	// Validate unique IDs
	if err := r.validateUniqueIDs(&dbCopy); err != nil {
		return 0, err
	}

	// Store database
	r.databaseMap[dbCopy.ID] = &dbCopy

	// Save to file
	if err := r.saveToFile(); err != nil {
		// Rollback on save failure
		delete(r.databaseMap, dbCopy.ID)
		return 0, fmt.Errorf("failed to save database: %w", err)
	}

	return dbCopy.ID, nil
}

func (r *DatabaseRepoMapImpl) GetDatabase(ID int) (Database, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	db, exists := r.databaseMap[ID]
	if !exists {
		return Database{}, fmt.Errorf("database with ID %d not found", ID)
	}

	// Return a copy to avoid external modifications
	dbCopy := *db
	dbCopy.Tables = make([]Table, len(db.Tables))
	for i, table := range db.Tables {
		dbCopy.Tables[i] = table
		dbCopy.Tables[i].Columns = make([]Column, len(table.Columns))
		copy(dbCopy.Tables[i].Columns, table.Columns)
	}

	return dbCopy, nil
}

func (r *DatabaseRepoMapImpl) GetAllDatabases() ([]Database, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	databases := make([]Database, 0, len(r.databaseMap))
	for _, db := range r.databaseMap {
		// Return a copy to avoid external modifications
		dbCopy := *db
		dbCopy.Tables = make([]Table, len(db.Tables))
		for i, table := range db.Tables {
			dbCopy.Tables[i] = table
			dbCopy.Tables[i].Columns = make([]Column, len(table.Columns))
			copy(dbCopy.Tables[i].Columns, table.Columns)
		}
		databases = append(databases, dbCopy)
	}

	return databases, nil
}

func (r *DatabaseRepoMapImpl) SetDescription(dbID int, desc string, fieldID int, fieldType fieldType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	db, exists := r.databaseMap[dbID]
	if !exists {
		return fmt.Errorf("database with ID %d not found", dbID)
	}

	switch fieldType {
	case DatabaseFieldType:
		if db.ID != fieldID {
			return fmt.Errorf("database ID mismatch: expected %d, got %d", db.ID, fieldID)
		}
		db.Description = desc

	case TableFieldType:
		found := false
		for i := range db.Tables {
			if db.Tables[i].ID == fieldID {
				db.Tables[i].Description = desc
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("table with ID %d not found in database %d", fieldID, dbID)
		}

	case ColumnFieldType:
		found := false
		for i := range db.Tables {
			for j := range db.Tables[i].Columns {
				if db.Tables[i].Columns[j].ID == fieldID {
					db.Tables[i].Columns[j].Description = desc
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return fmt.Errorf("column with ID %d not found in database %d", fieldID, dbID)
		}

	default:
		return fmt.Errorf("unknown field type: %d", fieldType)
	}

	// Save to file
	if err := r.saveToFile(); err != nil {
		return fmt.Errorf("failed to save description: %w", err)
	}

	return nil
}
