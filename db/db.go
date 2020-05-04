package db

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type db struct {
	conn *pgx.Conn
}

func OpenDB(conf map[string]interface{}) (*db, error) {
	return OpenPostgresql(conf["host"].(string),
		conf["user"].(string),
		conf["password"].(string),
		conf["dbname"].(string))
}

func (db *db) CloseDB() {
	db.conn.Close(context.TODO())
}

func (db *db) Exec(sql string) (pgconn.CommandTag, error) {
	return db.conn.Exec(context.TODO(), sql)
}

func (db *db) DropTable(table_name string) {
	db.Exec("DROP TABLE IF EXISTS " + table_name + " CASCADE")
}

func (db *db) Begin() (*Tx, error) {
	tx, err := db.conn.Begin(context.TODO())
	if err == nil {
		return &Tx{tx}, nil
	} else {
		return nil, err
	}
}

func (db *db) HasTable(table_name string) bool {
	_, err := db.conn.Query(context.TODO(), "SELECT * from "+table_name+" limit 1")
	return err == nil
}

func (db *db) PrepareAndExec(query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.conn.Exec(context.TODO(), query, args...)
}

type Tx struct {
	pgx.Tx
}

func (tx *Tx) DropTable(table_name string) {
	tx.Exec("DROP TABLE IF EXISTS " + table_name)
}

func (tx *Tx) HasTable(table_name string) bool {
	_, err := tx.Query("SELECT * from " + table_name + " limit 1")
	return err == nil
}

func (tx *Tx) Exec(sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return tx.Tx.Exec(context.TODO(), sql, args...)
}

func (tx *Tx) Query(sql string, args ...interface{}) (pgx.Rows, error) {
	return tx.Tx.Query(context.TODO(), sql, args...)
}

func (tx *Tx) Rollback() error {
	return tx.Tx.Rollback(context.TODO())
}

func (tx *Tx) Commit() error {
	return tx.Tx.Commit(context.TODO())
}
