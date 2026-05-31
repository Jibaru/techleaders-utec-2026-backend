// Package customer holds the JSON view shapes for the customer resource.
package customer

type CreateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
}
