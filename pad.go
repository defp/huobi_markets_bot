package main

func times(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}

// Right right-pads the string with pad up to len runes
func padRight(str string, length int, pad string) string {
	return str + times(pad, length-len(str))
}
