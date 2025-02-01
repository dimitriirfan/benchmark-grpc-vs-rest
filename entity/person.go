package entity

type GetPopulationResponse struct {
	Population []Person `json:"population"`
}

type Person struct {
	ID           string                 `json:"id"`
	FirstName    string                 `json:"first_name"`
	LastName     string                 `json:"last_name"`
	Email        string                 `json:"email"`
	DateOfBirth  string                 `json:"date_of_birth"`
	PhoneNumber  string                 `json:"phone_number"`
	Address      Address                `json:"address"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
	Active       bool                   `json:"active"`
	Role         string                 `json:"role"`
	ProfileImage string                 `json:"profile_image"`
	Preferences  map[string]interface{} `json:"preferences"`
}

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
}
