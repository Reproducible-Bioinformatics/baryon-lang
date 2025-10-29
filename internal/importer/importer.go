package importer

// Reads a string and imports its content, to later be processed to
// a bala program.
type Importer interface {
	Import(content []byte) error
	Export() (string, error)
}
