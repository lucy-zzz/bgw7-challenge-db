// go: generate mockgen -source=../customer_mysql.go-destination=mocks/customer_mock.go~package=mocks
package repository

import (
	"database/sql"
	"errors"

	"app/internal"
)

// NewCustomersMySQL creates new mysql repository for customer entity.
func NewCustomersMySQL(db *sql.DB) *CustomersMySQL {
	return &CustomersMySQL{db}
}

// CustomersMySQL is the MySQL repository implementation for customer entity.
type CustomersMySQL struct {
	// db is the database connection.
	db *sql.DB
}

// FindAll returns all customers from the database.
func (r *CustomersMySQL) FindAll() (c []internal.Customer, err error) {
	// execute the query
	rows, err := r.db.Query("SELECT `id`, `first_name`, `last_name`, `condition` FROM customers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// iterate over the rows
	for rows.Next() {
		var cs internal.Customer
		// scan the row into the customer
		err := rows.Scan(&cs.Id, &cs.FirstName, &cs.LastName, &cs.Condition)
		if err != nil {
			return nil, err
		}
		// append the customer to the slice
		c = append(c, cs)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	return
}

// Save saves the customer into the database.
func (r *CustomersMySQL) Save(c *internal.Customer) (err error) {
	// execute the query
	res, err := r.db.Exec(
		"INSERT INTO customers (`first_name`, `last_name`, `condition`) VALUES (?, ?, ?)",
		(*c).FirstName, (*c).LastName, (*c).Condition,
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
	(*c).Id = int(id)

	return
}

func (r *CustomersMySQL) GetTotalByCondition() ([]internal.TotalByConditionResponse, error) {
	query := "SELECT `condition`, ROUND(SUM(i.total), 2) AS total_rounded FROM customers c LEFT JOIN invoices i ON c.id = i.customer_id GROUP BY `condition`"
	rows, err := r.db.Query(query)

	if err != nil {
		return nil, errors.New("INTERNAL")
	}
	defer rows.Close()

	var results []internal.TotalByConditionResponse
	for rows.Next() {
		var r internal.TotalByConditionResponse
		if err := rows.Scan(&r.Condition, &r.TotalRounded); err != nil {
			return nil, errors.New("INTERNAL")
		}
		results = append(results, r)
	}

	return results, nil
}

func (r *CustomersMySQL) GetTopActiveCustomers() ([]internal.CustomerTopSpentResponse, error) {
	rows, err := r.db.Query(`
SELECT 
	c.id,
	c.first_name,
	c.last_name,
	SUM(i.total) AS total_spent
FROM 
	customers c
JOIN 
	invoices i ON c.id = i.customer_id
WHERE 
	c.condition = 1
GROUP BY 
	c.id, c.first_name, c.last_name
ORDER BY 
	total_spent DESC
LIMIT 5
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []internal.CustomerTopSpentResponse
	for rows.Next() {
		var r internal.CustomerTopSpentResponse
		if err := rows.Scan(&r.Id, &r.FirstName, &r.LastName, &r.TotalSpent); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, nil
}
