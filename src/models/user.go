package models

type UserResponse struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `	json:"data,omitempty"`
}
