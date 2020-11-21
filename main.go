package main

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func getRowCustomer(customerId int) (Customer, error) {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return Customer{}, err
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customer WHERE id=$1")
	if err != nil {
		return Customer{}, err
	}

	rows := stmt.QueryRow(customerId)

	customer := Customer{}
	err = rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Status)
	if err != nil {
		return Customer{}, err
	}

	return customer, nil

}

func removeCustomer(customerId int) error {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM customer WHERE id=$1")
	if err != nil {

		return err
	}

	_, err = stmt.Exec(customerId)

	// err = rows.Scan(&customerUpdated.ID, &customerUpdated.Name, &customerUpdated.Email, &customerUpdated.Status)
	if err != nil {
		return err
	}

	return nil

}
func updateCustomer(customerId int, customer Customer) (Customer, error) {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return Customer{}, err
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE customer SET name=$2 , email=$3, status=$4 WHERE id=$1;")
	if err != nil {
		return Customer{}, err
	}

	_, err = stmt.Exec(customerId, customer.Name, customer.Email, customer.Status)

	// err = rows.Scan(&customerUpdated.ID, &customerUpdated.Name, &customerUpdated.Email, &customerUpdated.Status)
	if err != nil {
		return Customer{}, err
	}

	customerUpdated := Customer{customerId, customer.Name, customer.Email, customer.Status}
	return customerUpdated, nil

}

func getCustomer() ([]Customer, error) {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return []Customer{}, err
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customer")
	if err != nil {
		return []Customer{}, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return []Customer{}, err
	}

	customers := []Customer{}
	for rows.Next() {
		customer := Customer{}
		err := rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Status)
		if err != nil {
			return []Customer{}, err
		}

		customers = append(customers, customer)

	}
	return customers, nil

}

func listCustomer(c *gin.Context) {
	var customer, err = getCustomer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, customer)
}

func listCustomerById(c *gin.Context) {
	customerId := c.Param("id")
	Id, err := strconv.Atoi(customerId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	customer, err := getRowCustomer(Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, customer)
}
func updateCustomerById(c *gin.Context) {
	customerId := c.Param("id")
	Id, err := strconv.Atoi(customerId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	customer := Customer{}
	err = c.ShouldBindJSON(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	customerUpdated, err := updateCustomer(Id, customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, customerUpdated)
}

func removeCustomerById(c *gin.Context) {
	customerId := c.Param("id")

	Id, err := strconv.Atoi(customerId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err = removeCustomer(Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "customer deleted"})
}

func createCustomer(c *gin.Context) {

	customer := Customer{}
	err := c.ShouldBindJSON(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	customerCreated, err := insertCustomer(customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, customerCreated)

}

func conn() {

}
func insertCustomer(customer Customer) (Customer, error) {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return Customer{}, err
	}
	defer db.Close()

	var query = `
	CREATE TABLE IF NOT EXISTS customer (
		id SERIAL PRIMARY KEY	 NOT NULL ,
		name text ,
		email text ,
		status text 
	)`
	_, err = db.Exec(query)
	if err != nil {
		return Customer{}, err
	}

	row := db.QueryRow("INSERT into customer(name, email, status) values($1, $2, $3) RETURNING id", customer.Name, customer.Email, customer.Status)

	var id int
	err = row.Scan(&id)
	if err != nil {
		return Customer{}, err
	}

	return Customer{id, customer.Name, customer.Email, customer.Status}, nil
}

func main() {
	r := gin.Default()
	r.POST("/customers", createCustomer)
	r.GET("/customers", listCustomer)
	r.GET("/customers/:id", listCustomerById)
	r.PUT("/customers/:id", updateCustomerById)
	r.DELETE("/customers/:id", removeCustomerById)
	r.Run(":2009")

}
