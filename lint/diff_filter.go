package lint

func normalizeChangedFilesSet(changedFiles []string) map[string]struct{} {
	if changedFiles == nil {
		return nil
	}

	set := make(map[string]struct{}, len(changedFiles))
	for _, file := range changedFiles {
		set[cleanPath(file)] = struct{}{}
	}
	return set
}

func filterInputFiles(inputFiles []string, changedFiles map[string]struct{}) []string {
	if changedFiles == nil {
		return inputFiles
	}

	filtered := make([]string, 0, len(inputFiles))
	for _, inputFile := range inputFiles {
		if _, ok := changedFiles[cleanPath(inputFile)]; ok {
			filtered = append(filtered, inputFile)
		}
	}
	return filtered
}
