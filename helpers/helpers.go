package helpers

func StringJoin(elements []string, delimiter string) string {
	result := ""
	for i, element := range elements {
		result += element
		if i < len(elements)-1 {
			result += delimiter
		}
	}
	return result
}
