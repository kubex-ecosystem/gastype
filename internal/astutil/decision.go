package astutil

// DeveConverterBools decide se vale a pena converter essa struct para bitflags
func DeveConverterBools(numBools int) bool {
	// Regra global inicial: só converte se tiver 3 ou mais bools
	return numBools >= 3
}
