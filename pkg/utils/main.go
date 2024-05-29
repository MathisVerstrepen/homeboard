package utils

import (
	"fmt"
)

func BytesToReadable(bytes int) string {
	const (
		KiB = 1024
		MiB = 1024 * KiB
		GiB = 1024 * MiB
		TiB = 1024 * GiB
	)

	switch {
	case bytes >= TiB:
		return fmt.Sprintf("%.1fTiB", float64(bytes)/TiB)
	case bytes >= GiB:
		return fmt.Sprintf("%.1fGiB", float64(bytes)/GiB)
	case bytes >= MiB:
		return fmt.Sprintf("%.1fMiB", float64(bytes)/MiB)
	case bytes >= KiB:
		return fmt.Sprintf("%.1fKiB", float64(bytes)/KiB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
