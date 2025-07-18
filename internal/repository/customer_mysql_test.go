package repository_test

import (
	"app/internal/repository"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-txdb"
	_ "github.com/DATA-DOG/go-txdb"

	"github.com/stretchr/testify/assert"
)

func TestCustomerRepo(t *testing.T) {
	txdb.Register("txdb", "mysql", "root:SENTINELA!@/test_bgw7")
	db, err := sql.Open("txdb", "testCustomerRepo")
	assert.NoError(t, err)
	defer db.Close()
	_, err = db.Exec("DROP TABLE IF EXISTS invoices;")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("DROP TABLE IF EXISTS customers;")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE customers (id INT PRIMARY KEY, `condition` INT);")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE invoices (id INT PRIMARY KEY, customer_id INT, total FLOAT);")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO customers (id, `condition`) VALUES (1, 10), (2, 20);")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO invoices (id, customer_id, total) VALUES (1,1,100.25), (2,2,200.75), (3,1,50.00);`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("should test GetTotalByCondition func - success", func(t *testing.T) {
		repo := repository.NewCustomersMySQL(db)

		results, err := repo.GetTotalByCondition()
		if err != nil {
			t.Fatal(err)
		}

		expected := map[int]float64{
			10: 150.25,
			20: 200.75,
		}

		assert.Equal(t, len(expected), len(results))
	})
}
