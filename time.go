package minicommerce

// TimeService is an abstraction on top of time pkg
// for easier testability whenever time is used in the minicommerce
type TimeService interface {
	Now() int64
}
