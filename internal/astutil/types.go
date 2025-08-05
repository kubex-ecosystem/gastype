package astutil

// MenorTipoParaFlags retorna o tipo de inteiro necessário para armazenar as flags
func MenorTipoParaFlags(numFlags int) string {
	switch {
	case numFlags <= 8:
		return "uint8"
	case numFlags <= 16:
		return "uint16"
	case numFlags <= 32:
		return "uint32"
	default:
		return "uint64"
	}
}

// GetConstNames returns the constant names for fields
func GetConstNames(fields []string) []string {
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = "Flag" + field
	}
	return names
}
