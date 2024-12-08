package types

type User interface {
	// ID() string
	Key() string
	GetToken() string
}
