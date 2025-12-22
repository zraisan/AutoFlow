package registry

type ExtractorResult struct {
	Runtime        string
	RuntimeVersion string
	PackageManager string
	Scripts        map[string]string
}

type Extractor interface {
	Name() string
	Extract(path string) (*ExtractorResult, error)
}

type Executor interface {
	Name() string
	Generate(result *ExtractorResult, path, name string) (string, error)
}
