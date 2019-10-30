package cloudlogging

// getArg returns the n:th value in args, if defined, otherwise returns false
// in the boolean return value. n must be a positive value.
//
// Use it as such:
//  if value, ok := getArg(1, args); ok {
//    // use 'value'
//  }
func getArg(n int, args ...string) (string, bool) {
	if n < 0 {
		return "", false
	}

	if len(args) > n {
		return args[n], true
	}

	return "", false
}
