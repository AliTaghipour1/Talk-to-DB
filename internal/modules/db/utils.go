package db

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// SQLValueToGo converts a SQL column value to an appropriate Go type
func SQLValueToGo(columnType *sql.ColumnType, value interface{}) string {
	if value == nil {
		return "nil"
	}

	// Get the database type name
	dbType := columnType.DatabaseTypeName()

	// Get the scan type (the Go type the driver will use)
	v := columnType.ScanType()

	// Handle common cases based on database type name
	switch dbType {
	// String types
	case "CHAR", "VARCHAR", "TEXT", "CLOB", "LONGTEXT", "MEDIUMTEXT", "TINYTEXT":
		switch val := value.(type) {
		case []byte:
			return string(val)
		case string:
			return val
		default:
			return fmt.Sprintf("%v", val)
		}

	// Integer types
	case "INT", "INTEGER", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT":
		num, ok := handleIntegerType(value)
		if !ok {
			return fmt.Sprintf("%v", value)
		}
		return fmt.Sprintf("%d", num)

	// Unsigned integer types
	case "UNSIGNED INT", "UNSIGNED INTEGER", "UNSIGNED BIGINT":
		num, ok := handleUnsignedIntegerType(value)
		if !ok {
			return fmt.Sprintf("%v", value)
		}
		return fmt.Sprintf("%d", num)

	// Float types
	case "FLOAT", "DOUBLE", "DECIMAL", "NUMERIC", "REAL":
		num, ok := handleFloatType(value)
		if !ok {
			return fmt.Sprintf("%v", value)
		}
		return fmt.Sprintf("%f", num)

	// Boolean types
	case "BOOL", "BOOLEAN", "BIT":
		result, ok := handleBooleanType(value)
		if !ok {
			return fmt.Sprintf("%v", value)
		}
		return fmt.Sprintf("%t", result)

	// Date/Time types
	case "DATE", "TIME", "DATETIME", "TIMESTAMP", "YEAR":
		date, ok := handleDateType(value)
		if !ok {
			return fmt.Sprintf("%v", value)
		}
		return date.Format(time.RFC3339)

	// Binary types
	case "BINARY", "VARBINARY", "BLOB", "LONGBLOB", "MEDIUMBLOB", "TINYBLOB", "BYTEA":
		switch val := value.(type) {
		case []byte:
			// Return as base64 encoded string for JSON compatibility
			return base64.StdEncoding.EncodeToString(val)
		default:
			return fmt.Sprintf("%v", value)
		}

	// JSON types
	case "JSON", "JSONB":
		switch val := value.(type) {
		case []byte:
			// Return as string - let the JSON marshaler handle it
			return string(val)
		case string:
			return val
		default:
			return fmt.Sprintf("%v", value)
		}

	// UUID types
	case "UUID", "UNIQUEIDENTIFIER":
		switch val := value.(type) {
		case []byte:
			if len(val) == 16 {
				return fmt.Sprintf("%x-%x-%x-%x-%x",
					val[0:4], val[4:6], val[6:8], val[8:10], val[10:16])
			}
			return string(val)
		case string:
			return val
		default:
			return fmt.Sprintf("%v", val)
		}

	// Array types (PostgreSQL)
	case "_INT4", "_INT8", "_TEXT", "_VARCHAR", "_BOOL", "_FLOAT4", "_FLOAT8":
		// Handle PostgreSQL array types
		return fmt.Sprintf("%v", value) // The driver usually returns a slice

	// Geometry types
	case "POINT", "LINESTRING", "POLYGON", "GEOMETRY":
		// Return as WKT (Well-Known Text) or GeoJSON
		switch val := value.(type) {
		case []byte:
			return string(val)
		default:
			return fmt.Sprintf("%v", val)
		}
	}

	valueData := reflect.Indirect(reflect.ValueOf(value))
	// Fallback: use reflection to determine type based on the actual value
	switch v.Kind() {
	case reflect.Slice, reflect.Array:

		if v.Elem().Kind() == reflect.Uint8 {
			// []byte -> try to convert to string if printable
			byteSlice := valueData.Bytes()
			if isPrintableASCII(byteSlice) {
				return string(byteSlice)
			}
			// Otherwise base64 encode
			return base64.StdEncoding.EncodeToString(byteSlice)
		}
		return fmt.Sprintf("%v", value)

	case reflect.Struct:
		// Check if it's a time.Time
		if t, ok := value.(time.Time); ok {
			return t.Format(time.RFC3339)
		}
		return fmt.Sprintf("%v", value)

	case reflect.Ptr:
		// Dereference pointer
		if valueData.IsNil() {
			return "nil"
		}
		return fmt.Sprintf("%v", valueData.Elem().Interface())

	default:
		return fmt.Sprintf("%v", value)
	}
}

func handleIntegerType(value interface{}) (int64, bool) {
	switch val := value.(type) {
	case int64:
		return val, true
	case int32:
		return int64(val), true
	case int:
		return int64(val), true
	case []byte:
		i, _ := strconv.ParseInt(string(val), 10, 64)
		return i, true
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i, true
	default:
		return 0, false
	}
}

func handleUnsignedIntegerType(value interface{}) (uint64, bool) {
	switch val := value.(type) {
	case uint64:
		return val, true
	case uint32:
		return uint64(val), true
	case uint:
		return uint64(val), true
	case []byte:
		u, _ := strconv.ParseUint(string(val), 10, 64)
		return u, true
	case string:
		u, _ := strconv.ParseUint(val, 10, 64)
		return u, true
	default:
		return 0, false
	}
}

func handleFloatType(value interface{}) (float64, bool) {
	switch val := value.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case []byte:
		f, _ := strconv.ParseFloat(string(val), 64)
		return f, true
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f, true
	default:
		return 0, false
	}
}

func handleBooleanType(value interface{}) (bool, bool) {
	switch val := value.(type) {
	case bool:
		return val, true
	case int64:
		return val != 0, true
	case []byte:
		if string(val) == "1" || string(val) == "true" {
			return true, true
		}
		return false, true
	case string:
		if val == "1" || val == "true" {
			return true, true
		}
		return false, true
	default:
		return false, false
	}
}

func handleDateType(value interface{}) (time.Time, bool) {
	switch val := value.(type) {
	case time.Time:
		return val, true
	case []byte:
		t, _ := time.Parse("2006-01-02 15:04:05", string(val))
		return t, true
	case string:
		t, _ := time.Parse("2006-01-02 15:04:05", val)
		return t, true
	default:
		return time.Time{}, false
	}
}

// Helper function to check if byte slice contains only printable ASCII
func isPrintableASCII(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}
