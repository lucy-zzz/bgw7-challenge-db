package application

import (
	"app/internal/handler"
	"app/internal/repository"
	"app/internal/service"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-sql-driver/mysql"
)

// ConfigApplicationDefault is the configuration for NewApplicationDefault.
type ConfigApplicationDefault struct {
	// Db is the database configuration.
	Db *mysql.Config
	// Addr is the server address.
	Addr string
}

// NewApplicationDefault creates a new ApplicationDefault.
func NewApplicationDefault(config *ConfigApplicationDefault) *ApplicationDefault {
	// default values
	defaultCfg := &ConfigApplicationDefault{
		Db:   nil,
		Addr: ":8080",
	}
	if config != nil {
		if config.Db != nil {
			defaultCfg.Db = config.Db
		}
		if config.Addr != "" {
			defaultCfg.Addr = config.Addr
		}
	}

	return &ApplicationDefault{
		cfgDb:   defaultCfg.Db,
		cfgAddr: defaultCfg.Addr,
	}
}

// ApplicationDefault is an implementation of the Application interface.
type ApplicationDefault struct {
	// cfgDb is the database configuration.
	cfgDb *mysql.Config
	// cfgAddr is the server address.
	cfgAddr string
	// db is the database connection.
	db *sql.DB
	// router is the chi router.
	router *chi.Mux
}

// SetUp sets up the application.
func (a *ApplicationDefault) SetUp() (err error) {
	if err := a.createDatabaseIfNotExists(); err != nil {
		return err
	}

	a.db, err = sql.Open("mysql", a.cfgDb.FormatDSN())
	if err != nil {
		return fmt.Errorf("failed to connect to db fantasy_products: %w", err)
	}

	if err := a.db.Ping(); err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}

	if err := a.createTablesIfNotExist(); err != nil {
		return err
	}

	// - repository
	rpCustomer := repository.NewCustomersMySQL(a.db)
	rpProduct := repository.NewProductsMySQL(a.db)
	rpInvoice := repository.NewInvoicesMySQL(a.db)
	rpSale := repository.NewSalesMySQL(a.db)
	// - service
	svCustomer := service.NewCustomersDefault(rpCustomer)
	svProduct := service.NewProductsDefault(rpProduct)
	svInvoice := service.NewInvoicesDefault(rpInvoice)
	svSale := service.NewSalesDefault(rpSale)
	// - handler
	hdCustomer := handler.NewCustomersDefault(svCustomer)
	hdProduct := handler.NewProductsDefault(svProduct)
	hdInvoice := handler.NewInvoicesDefault(svInvoice)
	hdSale := handler.NewSalesDefault(svSale)

	// routes
	// - router
	a.router = chi.NewRouter()
	// - middlewares
	a.router.Use(middleware.Logger)
	a.router.Use(middleware.Recoverer)
	// - endpoints
	a.router.Route("/customers", func(r chi.Router) {
		// - GET /customers
		r.Get("/", hdCustomer.GetAll())
		// - POST /customers
		r.Post("/", hdCustomer.Create())
		r.Get("/getTotalsByCondition", hdCustomer.GetTotalsByCondition())
		r.Get("/getTopActiveCustomers", hdCustomer.GetTopActiveCustomers())
	})
	a.router.Route("/products", func(r chi.Router) {
		// - GET /products
		r.Get("/", hdProduct.GetAll())
		// - POST /products
		r.Post("/", hdProduct.Create())
		r.Get("/getTopProductsSold", hdProduct.GetTopProducts())
	})
	a.router.Route("/invoices", func(r chi.Router) {

		r.Patch("/{id}", hdInvoice.Update())
		// - GET /invoices
		r.Get("/", hdInvoice.GetAll())
		// - POST /invoices
		r.Post("/", hdInvoice.Create())
	})
	a.router.Route("/sales", func(r chi.Router) {
		// - GET /sales
		r.Get("/", hdSale.GetAll())
		// - POST /sales
		r.Post("/", hdSale.Create())
	})

	return
}

// Run runs the application.
func (a *ApplicationDefault) Run() (err error) {
	defer a.db.Close()

	err = http.ListenAndServe(a.cfgAddr, a.router)
	return
}

func (a *ApplicationDefault) createDatabaseIfNotExists() error {
	cfgNoDB := *a.cfgDb
	cfgNoDB.DBName = ""

	dbRoot, err := sql.Open("mysql", cfgNoDB.FormatDSN())
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL root: %w", err)
	}
	defer dbRoot.Close()

	_, err = dbRoot.Exec("DROP DATABASE fantasy_products")
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	_, err = dbRoot.Exec("CREATE DATABASE IF NOT EXISTS fantasy_products")
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}

func (a *ApplicationDefault) createTablesIfNotExist() error {
	queries := []string{
		// customers
		`CREATE TABLE IF NOT EXISTS customers (
			id int NOT NULL AUTO_INCREMENT,
			first_name varchar(45) DEFAULT NULL,
			last_name varchar(45) DEFAULT NULL,
			` + "`condition`" + ` tinyint(1) DEFAULT NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		// invoices
		`CREATE TABLE IF NOT EXISTS invoices (
			id int NOT NULL AUTO_INCREMENT,
			datetime datetime DEFAULT NULL,
			customer_id int DEFAULT NULL,
			total float DEFAULT NULL,
			PRIMARY KEY (id),
			KEY idx_invoices_customer_id (customer_id),
			CONSTRAINT fk_invoices_customer_id FOREIGN KEY (customer_id) REFERENCES customers (id) ON DELETE CASCADE ON UPDATE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		// products
		`CREATE TABLE IF NOT EXISTS products (
			id int NOT NULL AUTO_INCREMENT,
			description varchar(100) DEFAULT NULL,
			price float DEFAULT NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		// sales
		`CREATE TABLE IF NOT EXISTS sales (
			id int NOT NULL AUTO_INCREMENT,
			quantity int DEFAULT NULL,
			invoice_id int DEFAULT NULL,
			product_id int DEFAULT NULL,
			PRIMARY KEY (id),
			KEY idx_sales_invoice_id (invoice_id),
			KEY idx_sales_product_id (product_id),
			CONSTRAINT fk_sales_invoice_id FOREIGN KEY (invoice_id) REFERENCES invoices (id) ON DELETE CASCADE ON UPDATE CASCADE,
			CONSTRAINT fk_sales_product_id FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE ON UPDATE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}

	for i, q := range queries {
		if _, err := a.db.Exec(q); err != nil {
			return fmt.Errorf("error creating table #%d: %w", i+1, err)
		}
	}

	return nil
}

func (a *ApplicationDefault) InsertCustomersJSON() error {
	data, err := os.ReadFile("docs/db/json/customers.json")
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	var customers []handler.CustomerJSON
	if err := json.Unmarshal(data, &customers); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	stmt, err := a.db.Prepare("INSERT INTO customers (id, first_name, last_name, `condition`) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, c := range customers {
		_, err = stmt.Exec(c.Id, c.FirstName, c.LastName, c.Condition)
		if err != nil {
			log.Printf("Error inserting customer ID %d: %v", c.Id, err)
			return err
		} else {
			fmt.Printf("Inserted customer: %v\n", c)
		}
	}

	return nil
}

func (a *ApplicationDefault) InsertInvoicesJSON() error {
	data, err := os.ReadFile("docs/db/json/invoices.json")
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	var invoices []handler.InvoiceJSON
	if err := json.Unmarshal(data, &invoices); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	stmt, err := a.db.Prepare("INSERT INTO invoices (id, datetime, customer_id, total) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, c := range invoices {
		_, err = stmt.Exec(c.Id, c.Datetime, c.CustomerId, c.Total)
		if err != nil {
			log.Printf("Error inserting invoice ID %d: %v", c.Id, err)
			return err
		} else {
			fmt.Printf("Inserted invoice: %v\n", c)
		}
	}

	return nil
}

func (a *ApplicationDefault) InsertProductsJSON() error {
	data, err := os.ReadFile("docs/db/json/products.json")
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	var products []handler.ProductJSON
	if err := json.Unmarshal(data, &products); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	stmt, err := a.db.Prepare("INSERT INTO products (id, description, price) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, p := range products {
		_, err = stmt.Exec(p.Id, p.Description, p.Price)
		if err != nil {
			log.Printf("Error inserting product ID %d: %v", p.Id, err)
			return err
		} else {
			fmt.Printf("Inserted product: %v\n", p)
		}
	}

	return nil
}

func (a *ApplicationDefault) InsertSalesJSON() error {
	data, err := os.ReadFile("docs/db/json/sales.json")
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	var sales []handler.SaleJSON
	if err := json.Unmarshal(data, &sales); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	stmt, err := a.db.Prepare("INSERT INTO sales (id, quantity, invoice_id, product_id) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, s := range sales {
		_, err = stmt.Exec(s.Id, s.Quantity, s.InvoiceId, s.ProductId)
		if err != nil {
			log.Printf("Error inserting sale ID %d: %v", s.Id, err)
			return err
		} else {
			fmt.Printf("Inserted sale: %v\n", s)
		}
	}

	return nil
}
