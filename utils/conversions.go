package utils

import "strconv"

// StringToUint convierte un string a uint de forma segura (devuelve 0 si falla)
func StringToUint(s string) uint {
	val, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint(val)
}

// UintToString convierte un uint a string de forma segura (devuelve "" si falla)
func UintToString(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}
