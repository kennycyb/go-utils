package strutil

// IsEmpty true if given string is nil or empty
func IsEmpty(s *string) bool {
	return s == nil || len(*s) == 0
}
