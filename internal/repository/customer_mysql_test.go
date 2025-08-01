package repository_test

import (
	"app/internal/repository"
	repository_test "app/internal/repository/test"
	"database/sql"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var customerOnce sync.Once

func setupCustomerTestDB(t *testing.T) *sql.DB {
	repository_test.RegisterTxDB()

	db, err := sql.Open("txdb", t.Name())
	require.NoError(t, err)

	queries := []string{
		"DROP TABLE IF EXISTS invoices;",
		"DROP TABLE IF EXISTS customers;",
		"CREATE TABLE customers (id INT PRIMARY KEY AUTO_INCREMENT, first_name VARCHAR(255), last_name VARCHAR(255), `condition` INT);",
		"CREATE TABLE invoices (id INT PRIMARY KEY AUTO_INCREMENT, customer_id INT, total FLOAT);",
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		require.NoError(t, err)
	}

	return db
}

func insertCustomerData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("INSERT INTO customers (first_name, last_name, `condition`) VALUES ('Fulano', 'de Tal', 10), ('Fulanete', 'Zete', 20);")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO invoices (customer_id, total) VALUES (1, 100.25), (2, 200.75), (1, 50.00);")
	require.NoError(t, err)
}

func TestCustomersMySQL_GetTotalByCondition(t *testing.T) {
	db := setupCustomerTestDB(t)
	defer db.Close()

	insertCustomerData(t, db)

	repo := repository.NewCustomersMySQL(db)

	t.Run("success - total by condition", func(t *testing.T) {
		results, err := repo.GetTotalByCondition()
		require.NoError(t, err)

		expected := map[int]float64{
			10: 150.25,
			20: 200.75,
		}

		assert.Len(t, results, len(expected))

		for _, r := range results {
			expectedTotal, ok := expected[r.Condition]
			require.True(t, ok)
			assert.InDelta(t, expectedTotal, r.TotalRounded, 0.01)
		}
	})

	t.Run("fail - simulate db error by closing db", func(t *testing.T) {
		db.Close()
		results, err := repo.GetTotalByCondition()
		assert.Nil(t, results)
		assert.Error(t, err)
	})
}

func TestCustomersMySQL_GetTopActiveCustomers(t *testing.T) {
	db := setupCustomerTestDB(t)
	defer db.Close()

	_, err := db.Exec("INSERT INTO customers (first_name, last_name, `condition`) VALUES ('Ana', 'A', 1), ('Beto', 'B', 0), ('Cida', 'C', 1);")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO invoices (customer_id, total) VALUES (1, 100), (1, 200), (3, 300);")
	require.NoError(t, err)

	repo := repository.NewCustomersMySQL(db)

	t.Run("success - top active customers", func(t *testing.T) {
		results, err := repo.GetTopActiveCustomers()
		require.NoError(t, err)

		assert.Len(t, results, 2)

		if len(results) >= 2 {
			assert.GreaterOrEqual(t, results[0].TotalSpent, results[1].TotalSpent)
		}
	})

	t.Run("fail - db closed", func(t *testing.T) {
		db.Close()
		results, err := repo.GetTopActiveCustomers()
		assert.Nil(t, results)
		assert.Error(t, err)
	})
}
