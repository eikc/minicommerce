package minicommerce

// IDGenerator is the abstraction for generating ID's in minicommerce
type IDGenerator interface {
	New() (string, error)
}
