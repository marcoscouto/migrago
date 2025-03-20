package transaction

import "database/sql"

type Transaction interface {
	Execute(func(transaction *sql.Tx) error) error
}

type transaction struct {
	conn *sql.DB
}

func New(conn *sql.DB) Transaction {
	return &transaction{
		conn: conn,
	}
}

func (t *transaction) Execute(f func(tx *sql.Tx) error) error {
	tx, err := t.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := f(tx); err != nil {
		return err
	}

	return tx.Commit()
}
