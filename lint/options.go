package lint

import "runtime"

const defaultMaxLintConcurrency = 4

func effectiveLintConcurrency(ruleCount int) int {
	if ruleCount <= 0 {
		return 1
	}

	cfg := getConfig()
	if cfg != nil && cfg.Lint.Concurrency != nil && *cfg.Lint.Concurrency > 0 {
		if *cfg.Lint.Concurrency > ruleCount {
			return ruleCount
		}
		return *cfg.Lint.Concurrency
	}

	auto := runtime.GOMAXPROCS(0)
	if auto < 1 {
		auto = 1
	}
	if auto > defaultMaxLintConcurrency {
		auto = defaultMaxLintConcurrency
	}
	if auto > ruleCount {
		auto = ruleCount
	}
	return auto
}

func regoTraceEnabled() bool {
	cfg := getConfig()
	return cfg != nil && cfg.Lint.RegoTrace != nil && *cfg.Lint.RegoTrace
}
