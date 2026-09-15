package mpr

import "runtime"

const defaultMaxExportConcurrency = 4

var configuredExportConcurrency int

func SetExportConcurrency(concurrency int) {
	if concurrency < 0 {
		concurrency = 0
	}
	configuredExportConcurrency = concurrency
}

// ConfigureExportConcurrency applies export.concurrency from config. Zero or nil uses GOMAXPROCS-based default.
func ConfigureExportConcurrency(concurrency *int) {
	if concurrency == nil || *concurrency <= 0 {
		SetExportConcurrency(0)
		return
	}
	SetExportConcurrency(*concurrency)
}

func effectiveExportConcurrency(documentCount int) int {
	if documentCount <= 0 {
		return 1
	}

	if configuredExportConcurrency > 0 {
		if configuredExportConcurrency > documentCount {
			return documentCount
		}
		return configuredExportConcurrency
	}

	auto := runtime.GOMAXPROCS(0)
	if auto < 1 {
		auto = 1
	}
	if auto > defaultMaxExportConcurrency {
		auto = defaultMaxExportConcurrency
	}
	if auto > documentCount {
		auto = documentCount
	}
	return auto
}
