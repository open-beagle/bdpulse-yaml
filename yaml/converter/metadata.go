package converter

// Metadata provides additional metadata used to
// convert the configuration file format.
type Metadata struct {
	// Filename of the configuration file, helps
	// determine the yaml configuration format.
	Filename string

	// URL of the repository used to create the repository
	// workspace directory using the fully qualified name.
	// e.g. /drone/src/github.com/octocat/hello-world
	URL string

	// Ref of the commit used to choose the correct
	// pipeline if the configuration format defines
	// multiple pipelines (like Bitbucket)
	Ref string
}
