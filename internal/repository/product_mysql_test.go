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

var once sync.Once

func setupTestDBProducts(t *testing.T) *sql.DB {
	repository_test.RegisterTxDB()

	db, err := sql.Open("txdb", t.Name())
	require.NoError(t, err)

	queries := []string{
		"DROP TABLE IF EXISTS sales;",
		"DROP TABLE IF EXISTS products;",

		`CREATE TABLE products (
			id INT PRIMARY KEY AUTO_INCREMENT,
			description VARCHAR(255),
			price FLOAT
		);`,

		`CREATE TABLE sales (
			id INT PRIMARY KEY AUTO_INCREMENT,
			quantity INT,
			invoice_id INT,
			product_id INT
		);`,
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		require.NoError(t, err)
	}

	return db
}

func insertSampleDataProducts(t *testing.T, db *sql.DB) {
	_, err := db.Exec("INSERT INTO products (description, price) VALUES ('Produto A', 10), ('Produto B', 20), ('Produto C', 30);")
	require.NoError(t, err)

	_, err = db.Exec(`
	INSERT INTO sales (quantity, invoice_id, product_id) VALUES
	(5, 1, 1),
	(3, 2, 1),
	(10, 3, 2),
	(2, 4, 3);
	`)
	require.NoError(t, err)
}

func TestProductsMySQL_GetTopProducts(t *testing.T) {
	db := setupTestDBProducts(t)
	defer db.Close()

	insertSampleDataProducts(t, db)

	repo := repository.NewProductsMySQL(db)

	t.Run("success - get top products", func(t *testing.T) {
		results, err := repo.GetTopProducts()
		require.NoError(t, err)

		assert.Len(t, results, 3)

		if len(results) >= 2 {
			assert.GreaterOrEqual(t, results[0].TotalQuantitySold, results[1].TotalQuantitySold)
		}

		expected := map[int]int{
			1: 8,
			2: 10,
			3: 2,
		}
		for _, p := range results {
			expQuantity, ok := expected[p.Id]
			require.True(t, ok)
			assert.Equal(t, expQuantity, p.TotalQuantitySold)
		}
	})

	t.Run("fail - simulate db error closing connection", func(t *testing.T) {
		db.Close()
		results, err := repo.GetTopProducts()
		assert.Nil(t, results)
		assert.Error(t, err)
	})
}
