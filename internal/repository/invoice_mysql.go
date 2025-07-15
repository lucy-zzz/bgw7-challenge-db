package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"app/internal"

	"github.com/go-sql-driver/mysql"
)

// NewInvoicesMySQL creates new mysql repository for invoice entity.
func NewInvoicesMySQL(db *sql.DB) *InvoicesMySQL {
	return &InvoicesMySQL{db}
}

// InvoicesMySQL is the MySQL repository implementation for invoice entity.
type InvoicesMySQL struct {
	// db is the database connection.
	db *sql.DB
}

// FindAll returns all invoices from the database.
func (r *InvoicesMySQL) FindAll() (i []internal.Invoice, err error) {
	// execute the query
	rows, err := r.db.Query("SELECT `id`, `datetime`, `total`, `customer_id` FROM invoices")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// iterate over the rows
	for rows.Next() {
		var iv internal.Invoice
		// scan the row into the invoice
		err := rows.Scan(&iv.Id, &iv.Datetime, &iv.Total, &iv.CustomerId)
		if err != nil {
			return nil, err
		}
		// append the invoice to the slice
		i = append(i, iv)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	return
}

// Save saves the invoice into the database.
func (r *InvoicesMySQL) Save(i *internal.Invoice) (err error) {
	// execute the query
	res, err := r.db.Exec(
		"INSERT INTO invoices (`datetime`, `total`, `customer_id`) VALUES (?, ?, ?)",
		(*i).Datetime, (*i).Total, (*i).CustomerId,
	)
	if err != nil {
		return err
	}

	// get the last inserted id
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// set the id
	(*i).Id = int(id)

	return
}

func (r *InvoicesMySQL) FindById(id int) (iv internal.Invoice, err error) {
	rows, err := r.db.Query(
		"SELECT * FROM invoices WHERE id = ?",
		id,
	)

	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				return iv, mysqlErr
			}
		}
		return iv, err
	}

	for rows.Next() {
		err := rows.Scan(&iv.Id, &iv.Datetime, &iv.Total, &iv.CustomerId)
		if err != nil {
			return iv, err
		}
	}

	return iv, err
}

// func (r *InvoicesMySQL) Update(i internal.InvoiceAttributes, id int) (iv internal.Invoice, err error) {
// 	inv, err := r.FindById(id)

// 	if err != nil {
// 		return iv, errors.New("ID not found.")
// 	}

// 	var willUpdate int

// 	if inv.CustomerId != i.CustomerId {
// 		willUpdate++
// 	}

// 	if inv.Datetime != i.Datetime {
// 		willUpdate++
// 	}

// 	if inv.Total != i.Total {
// 		willUpdate++
// 	}

// 	if willUpdate != 0 {
// 		_, err := r.db.Exec(
// 			"UPDATE invoices SET column = ?, ?, ? WHERE id = ?",
// 			(i).Datetime, (i).Total, (i).CustomerId, id,
// 		)

// 		if err != nil {
// 			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
// 				if mysqlErr.Number == 1062 {
// 					return iv, mysqlErr
// 				}
// 			}
// 			return iv, err
// 		}

// 		return iv, err
// 	} else {
// 		return iv, errors.New("Nothing to update.")
// 	}
// }

// func (r *InvoicesMySQL) Update(i internal.InvoiceAttributes, id int) (iv internal.Invoice, err error) {
// 	inv, err := r.FindById(id)

// 	if err != nil {
// 		return iv, errors.New("ID not found.")
// 	}

// 	if i.CustomerId != 0 {
// 		inv.CustomerId = i.CustomerId
// 	}

// 	if i.Datetime != "" {
// 		inv.Datetime = i.Datetime
// 	}

// 	if i.Total != 0 {
// 		inv.Total = i.Total
// 	}

// 	query := "UPDATE invoices SET datetime = ?, customer_id = ?, total = ? WHERE id = ?"

// 	_, err = r.db.Exec(query, inv.Datetime, inv.CustomerId, inv.Total, id)
// 	if err != nil {
// 		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
// 			if mysqlErr.Number == 1062 {
// 				return iv, errors.New("CONFLICT")
// 			}
// 		}
// 		return iv, errors.New("INTERNAL")
// 	}

// 	return iv, nil
// }

func (r *InvoicesMySQL) Update(i internal.InvoiceAttributes, id int) (iv internal.Invoice, err error) {
	var setClauses []string
	var args []interface{}

	_, err = r.FindById(id)

	if err != nil {
		return iv, errors.New("ID not found.")
	}

	if i.Datetime != "" {
		setClauses = append(setClauses, "datetime = ?")
		args = append(args, i.Datetime)
	}

	if i.CustomerId != 0 {
		setClauses = append(setClauses, "customer_id = ?")
		args = append(args, i.CustomerId)
	}

	if i.Total != 0 {
		setClauses = append(setClauses, "total = ?")
		args = append(args, i.Total)
	}

	if len(setClauses) == 0 {
		return iv, errors.New("BAD REQUEST")
	}

	query := fmt.Sprintf("UPDATE invoices SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)

	res, err := r.db.Exec(query, args...)
	if err != nil {
		return iv, errors.New("INTERNAL")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return iv, errors.New("INTERNAL")
	}

	if rowsAffected == 0 {
		return iv, errors.New("INTERNAL")
	}

	row := r.db.QueryRow("SELECT * FROM invoices WHERE id = ?", id)
	err = row.Scan(
		&iv.Id,
		&iv.Datetime,
		&iv.CustomerId,
		&iv.Total,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return iv, errors.New("NOT FOUND")
		}
		return iv, errors.New("INTERNAL")
	}

	return iv, nil
}
