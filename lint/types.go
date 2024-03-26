package lint

type PolicyResult int

const (
	PolicyResultPass PolicyResult = iota
	PolicyResultFail
	PolicyResultUnknown
)
