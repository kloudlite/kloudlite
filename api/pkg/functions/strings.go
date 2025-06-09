package functions

func StringReverse(x string) string {
	runes := []rune(x)
	for i := 0; i < len(runes)/2; i += 1 {
		runes[i], runes[len(runes)-i-1] = runes[len(runes)-i-1], runes[i]
	}
	return string(runes)
}
