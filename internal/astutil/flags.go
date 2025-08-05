// Package astutil fornece funções utilitárias para manipulação de flags
package astutil

// SetFlag ativa um flag específico
func SetFlag(v interface{}, f uint64) {
	switch ptr := v.(type) {
	case *uint8:
		*ptr |= uint8(f)
	case *uint16:
		*ptr |= uint16(f)
	case *uint32:
		*ptr |= uint32(f)
	case *uint64:
		*ptr |= uint64(f)
	}
}

// ClearFlag desativa um flag específico
func ClearFlag(v interface{}, f uint64) {
	switch ptr := v.(type) {
	case *uint8:
		*ptr &^= uint8(f)
	case *uint16:
		*ptr &^= uint16(f)
	case *uint32:
		*ptr &^= uint32(f)
	case *uint64:
		*ptr &^= uint64(f)
	}
}

// ToggleFlag inverte um flag específico
func ToggleFlag(v interface{}, f uint64) {
	switch ptr := v.(type) {
	case *uint8:
		*ptr ^= uint8(f)
	case *uint16:
		*ptr ^= uint16(f)
	case *uint32:
		*ptr ^= uint32(f)
	case *uint64:
		*ptr ^= uint64(f)
	}
}

// HasFlag verifica se um flag específico está ativo
func HasFlag(v interface{}, f uint64) bool {
	switch val := v.(type) {
	case uint8:
		return val&uint8(f) != 0
	case uint16:
		return val&uint16(f) != 0
	case uint32:
		return val&uint32(f) != 0
	case uint64:
		return val&uint64(f) != 0
	}
	return false
}
