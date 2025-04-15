package utils

func SafeExec(action func(...interface{}), args ...interface{}) {
	if action != nil {
		action(args...)
	}
}
