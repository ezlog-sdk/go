package ezlog

import "runtime"

type Field struct {
	File   string
	Line   int
	Func   string
	Caller string
}

func GetField(skip int) *Field {
	pc, file, line, ok := runtime.Caller(skip)

	var method string
	detail := runtime.FuncForPC(pc)
	if ok && detail != nil {
		method = detail.Name()
	}

	pc, _, _, ok = runtime.Caller(skip + 1)
	var caller string
	detail = runtime.FuncForPC(pc)
	if ok && detail != nil {
		caller = detail.Name()
	}

	field := &Field{
		File:   file,
		Line:   line,
		Func:   method,
		Caller: caller,
	}

	return field
}
