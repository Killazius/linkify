package domain

type User struct {
	ID       string
	Email    string
	PassHash []byte
}
