package databases

import (
	"database/sql"
	"fmt"
	"context"
	"database/sql/driver"
	"log"
	"os"
	"github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// foreignKeyConnector handles the database connection with foreign keys enabled.
type foreignKeyConnector struct {
	dsn    string
	driver driver.Driver
}

// Connect opens a new connection and enables foreign keys.
func (c *foreignKeyConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.driver.Open(c.dsn)
	if err != nil {
		return nil, err
	}
	execer, ok := conn.(driver.ExecerContext)
	if !ok {
		return nil, fmt.Errorf("connection does not implement ExecerContext")
	}
	_, err = execer.ExecContext(ctx, "PRAGMA foreign_keys = ON;", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	return conn, nil
}

func (c *foreignKeyConnector) Driver() driver.Driver {
	return c.driver
}

// InitDB initializes the database, enabling foreign keys and executing SQL statements.
func InitDB(filepath string) {
	connector := &foreignKeyConnector{
		dsn:    filepath,
		driver: &sqlite3.SQLiteDriver{},
	}

	var err error
	DB = sql.OpenDB(connector)

	// Check if foreign keys are enabled
	var foreignKeysEnabled int
	if err := DB.QueryRow("PRAGMA foreign_keys;").Scan(&foreignKeysEnabled); err != nil {
		log.Fatal("Failed to check foreign keys: ", err)
	}
	if foreignKeysEnabled == 0 {
		log.Fatal("Foreign keys are not enabled in the database.")
	}

	// Read and execute the SQL file
	sqlfile, err := os.ReadFile("./databases/my.sql")
	if err != nil {
		log.Fatal("Read error: ", err)
	}

	_, err = DB.Exec(string(sqlfile))
	if err != nil {
		log.Fatal("Exec error: ", err)
	}

	// Insert initial values into the categories table
	_, _ = DB.Exec(`INSERT INTO categories (name) VALUES 
		('Sport'), 
		('Music'), 
		('Science'), 
		('Technology'), 
		('Culture');`)

	fmt.Println("Database initialized and queries executed successfully!")
}
