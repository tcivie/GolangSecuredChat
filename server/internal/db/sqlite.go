package db

import (
	"crypto/rsa"
	crypt "crypto/x509"
	"database/sql"
	"fmt"
	"os"
	"sync"
)

type Database struct {
	conn *sql.DB
}

var instance *Database
var once sync.Once

const UsersDBPath = "../../resources/db/users.db"
const createUsersTableSQL = `CREATE TABLE IF NOT EXISTS Users(
    username TEXT PRIMARY KEY,
    pubkey BLOB
)`

func CreateConnection(dbPath string) (*sql.DB, error) {
	// Check if the file exists
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		fmt.Println("Database file does not exist. Creating it...")
		// Create the file
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("error creating database file: %v", err)
		}
		file.Close()
	}

	// Open the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Ping the database to verify the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	fmt.Println("Database created/opened successfully")
	return db, nil
}

func GetDatabase() *Database {
	once.Do(func() {
		conn, err := CreateConnection(UsersDBPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to create database connection: %v", err))
		}
		_, err = conn.Exec(createUsersTableSQL)
		if err != nil {
			panic(fmt.Sprintf("Failed to create Users table: %v", err))
		}
		fmt.Println("Users table created successfully")
		instance = &Database{conn: conn}
	})
	return instance
}

func (db *Database) query(sqlStatement string) (*sql.Rows, error) {
	rows, err := db.conn.Query(sqlStatement)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (db *Database) GetUserPubKey(username string) (*rsa.PublicKey, error) {
	var err error
	var rows *sql.Rows
	var rsaBubKey *rsa.PublicKey
	rows, err = db.query(fmt.Sprintf("SELECT pubkey FROM Users WHERE username = '%s';", username))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pubkey []byte
		if err = rows.Scan(&pubkey); err != nil {
			return nil, err
		}
		rsaBubKey, err = crypt.ParsePKCS1PublicKey(pubkey)
		return rsaBubKey, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (db *Database) CreateNewUser(username string, pubkey []byte) error {
	stmt, err := db.conn.Prepare("INSERT OR IGNORE INTO Users(username, pubkey) VALUES(?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, pubkey)
	if err != nil {
		return fmt.Errorf("error executing insert: %v", err)
	}

	fmt.Println("User inserted successfully")
	return nil
}
