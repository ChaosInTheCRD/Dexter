package files

// Found is a single file which was found by a parser to have an image reference inside it
type Found struct {
	// Location is the filepath location where the certificate was found.
	Location string

	// Parser is the name of the parser which discovered the file.
	Parser Parser

	// References are the image references found in the file (key) and the new reference (value)
	References []string

	// NewReferences are the new image references templated into the file
	NewReferences []string
}

type Parser struct {
	// Name is the name of the Parser
	Name string

	// Lead is the emoji or identifier used in the logs to identify the used Parser
	Lead string

	// Parser is the Parser interface
	Parser parser
}
