package lib

func assert(v bool) {
	if !v {
		Panic("Assertion failed.")
	}
}

func Assert(v bool) {
	assert(v)
}