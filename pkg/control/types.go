package control

// ResponseWriter is the interface for writing command responses
type ResponseWriter interface {
	WriteLine(string) error
}

