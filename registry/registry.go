package registry

var (
	extractors []Extractor
	executors  []Executor
)

func RegisterExtractor(e Extractor) {
	extractors = append(extractors, e)
}

func RegisterExecutor(e Executor) {
	executors = append(executors, e)
}

func ExtractorNames() []string {
	names := make([]string, len(extractors))
	for i, e := range extractors {
		names[i] = e.Name()
	}
	return names
}

func ExecutorNames() []string {
	names := make([]string, len(executors))
	for i, e := range executors {
		names[i] = e.Name()
	}
	return names
}

func GetExtractor(index int) Extractor {
	return extractors[index]
}

func GetExecutor(index int) Executor {
	return executors[index]
}
