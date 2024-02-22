package repo

import (
	"database/sql"
	"fmt"
	"log"

	pb "github.com/ilovepitsa/protobufForTestCase"
)

type TransactionRepository struct {
	l  *log.Logger
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB, l *log.Logger) *TransactionRepository {
	return &TransactionRepository{db: db, l: l}
}

func getCurrencyString(currency pb.CurrencyType) string {
	switch currency {
	case pb.CurrencyType_CURRENCY_BTC:
		return "btc"
	case pb.CurrencyType_CURRENCY_EUR:
		return "eur"
	case pb.CurrencyType_CURRENCY_RUB:
		return "rub"
	case pb.CurrencyType_CURRENCY_USD:
		return "usd"
	case pb.CurrencyType_CURRENCY_USDT:
		return "usdt"
	}
	return ""
}

func getQuery(transaction *pb.Transaction) string {

	switch transaction.Action {
	case pb.ActionType_ACTION_ADD:
		return fmt.Sprintf("insert into transaction (customerid, num_invoice, currency, amount, action, statustrans) values (%v, '%s', '%s', %v, 'add', 'created');",
			transaction.Id,
			transaction.Number_Invoice,
			getCurrencyString(transaction.Currency),
			transaction.Amount)
	case pb.ActionType_ACTION_SUB:
		return fmt.Sprintf("insert into transaction (customerid, num_invoice, currency, amount, action,to_num_invoice, statustrans) values (%v, '%s', '%s', %v, 'sub', '%s', 'created');",
			transaction.Id,
			transaction.Number_Invoice,
			getCurrencyString(transaction.Currency),
			transaction.Amount,
			transaction.Number_InvoiceTo)
	}
	return ""
}

func (x *TransactionRepository) blockAccounts(transaction *pb.Transaction, trans *sql.Tx) error {
	_, err := trans.Exec(fmt.Sprintf("update accounts set status = false where num_invoice = '%s';", transaction.Number_Invoice))
	if err != nil {
		return err
	}

	if len(transaction.Number_InvoiceTo) != 0 {
		_, err := trans.Exec(fmt.Sprintf("update accounts set status = false where num_invoice = '%s';", transaction.Number_InvoiceTo))
		if err != nil {
			return err
		}
	}
	return nil
}

func (x *TransactionRepository) Add(transaction *pb.Transaction) error {
	trans, err := x.db.Begin()
	if err != nil {
		x.l.Println(err)
		trans.Rollback()
		return err
	}
	// x.l.Println(transaction.Action)
	// x.l.Println(getQuery(transaction))
	rows, err := trans.Query("select * from accounts;")

	for rows.Next() {
		col, err := rows.Columns()
		x.l.Println(col, err)
	}

	stmt, err := trans.Exec(getQuery(transaction))
	if err != nil {
		x.l.Println(err)
		trans.Rollback()
		return err
	}
	id, err := stmt.RowsAffected()
	if err != nil {
		x.l.Println(err)
		trans.Rollback()
		return err
	}
	x.l.Printf("Row affected: %v", id)

	err = x.blockAccounts(transaction, trans)
	if err != nil {
		x.l.Println(err)
		trans.Rollback()
		return err
	}
	rows, err = trans.Query("select max(id) from transaction;")

	if err != nil {
		x.l.Println(err)
		trans.Rollback()
		return err
	}
	for rows.Next() {
		err = rows.Scan(&transaction.Id)
		if err != nil {
			x.l.Println(err)
		}
	}
	transaction.Id = transaction.Id + 1

	trans.Commit()

	return nil
}

func (x *TransactionRepository) releaseAccounts(transaction *pb.Transaction, trans *sql.Tx) error {
	_, err := trans.Exec(fmt.Sprintf("update accounts set Status = true where num_invoice = '%s';", transaction.Number_Invoice))
	if err != nil {
		return err
	}
	if len(transaction.Number_InvoiceTo) != 0 {
		_, err := trans.Exec(fmt.Sprintf("update accounts set Status = true where num_invoice = '%s';", transaction.Number_InvoiceTo))
		if err != nil {
			return err
		}
	}
	return nil
}

func (x *TransactionRepository) SetError(transaction *pb.Transaction) {
	trans, err := x.db.Begin()
	if err != nil {
		x.l.Println(err)
		trans.Rollback()
		return
	}
	_, err = trans.Exec(fmt.Sprintf("update transaction set statustrans = 'error' where id = %v;", transaction.Id))
	if err != nil {
		x.l.Println(err)
		trans.Rollback()
		return
	}
	err = x.releaseAccounts(transaction, trans)
	if err != nil {
		x.l.Println(err)
		trans.Rollback()
		return
	}
	trans.Commit()
}
