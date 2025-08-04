package internal

// CustomerAttributes is the struct that represents the attributes of a customer.
type CustomerAttributes struct {
	// FirstName is the first name of the customer.
	FirstName string
	// LastName is the last name of the customer.
	LastName string
	// Condition is the condition of the customer.
	Condition int
}

// Customer is the struct that represents a customer.
type Customer struct {
	// Id is the unique identifier of the customer.
	Id int
	// CustomerAttributes is the attributes of the customer.
	CustomerAttributes
}

type TotalByConditionResponse struct {
	Condition    int     `json:"condition"`
	TotalRounded float64 `json:"total_rounded"`
}

type CustomerTopSpentResponse struct {
	Id         int     `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	TotalSpent float64 `json:"total_spent"`
}
