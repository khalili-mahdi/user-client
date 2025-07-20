package user

import (
	"fmt"
	"net/http"
	"time"
)

const (
	authKey = "token"
)

type CanCustomerRes struct {
	Allowed bool `json:"allowed"`
	UserID  int  `json:"userID"`
}

func (c *Client) CanCustomer(token string) (*CanCustomerRes, error) {
	url := fmt.Sprintf("%s/user/internal/auth/customer", c.internal)
	var u CanCustomerRes
	err := c.httpCall(url, http.MethodGet, map[string]string{authKey: token}, nil, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type PermissionRes struct {
	TraceID string `json:"traceID"`
	Allowed bool   `json:"allowed"`
	UserID  int    `json:"userID"`
}

func (c *Client) CanAdmin(token, scope, action string) (*CanCustomerRes, error) {
	url := fmt.Sprintf("%s/user/internal/auth/admin/%s/%s", c.internal, scope, action)

	var u CanCustomerRes
	err := c.httpCall(url, http.MethodGet, map[string]string{authKey: token}, nil, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type User struct {
	ID           int       `json:"id"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	KycLevel     int       `json:"kycLevel"`
	NationalCode string    `json:"nationalCode"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (c *Client) User(token string) (*User, error) {
	url := fmt.Sprintf("%s/user/internal/me", c.internal)
	var u User
	err := c.httpCall(url, http.MethodGet, map[string]string{"x-auth-id": token}, nil, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
