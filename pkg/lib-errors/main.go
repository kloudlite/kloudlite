package lib_errors

func HandleErr(e *error) {
	if r := recover(); r != nil {
		if y, ok := r.(error); ok {
			*e = y
		}
		return
	}
}

func Assert(condition bool, err error) {
	if !condition {
		panic(err)
	}
}

func AssertNoError(err error, msg error) {
	if err != nil {
		panic(msg)
	}
}
