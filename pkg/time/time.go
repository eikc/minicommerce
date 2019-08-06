package time

import (
	"time"
)

// Service is an abstraction over the time package for better testability
type Service struct {
}

// NewService will construct the Time.Service
func NewService() *Service {
	return &Service{}
}

// Now returns the time.Now UTC in unix format
func (s *Service) Now() int64 {
	return time.Now().UTC().Unix()
}
