package functions

func UpdateCodeVersion(code string) string {
	size := len(code)	
	char := []rune(code)[size-1]
	nextVersion := NextRune(char)
	res := code[:size-1] + string(nextVersion)
	return res
}

func NextRune(r rune) rune {
	return r + 1
}
