package main

// Type to store list of strings passed with repeating CLI flag
type stringListFlag []string

// Converter to string reqd to use type with flag package
func (s *stringListFlag) String() string {
	r := ""
	for _, v := range *s {
		if len(r) > 0 {
			r += ", "
		}
		r += v
	}
	return r
}

// Set Setter needed to use type with flag package
func (s *stringListFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}
