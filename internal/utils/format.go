package utils

import (
	"fmt"
	"math"
)

func FormatSizePrecise(bytes int64) string {
	const unit = 1024

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	exp := int(math.Log(float64(bytes)) / math.Log(unit))

	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

	if exp >= len(units) {
		exp = len(units) - 1
	}

	value := float64(bytes) / math.Pow(unit, float64(exp))

	precision := 2

	if value >= 100 {
		precision = 0
	} else if value >= 10 {
		precision = 1
	}

	return fmt.Sprintf("%.*f %s", precision, value, units[exp])
}
