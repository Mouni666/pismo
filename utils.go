package main

func normalizeAmount(opType int, amt float64) float64 {
	switch opType {
	case 1, 2, 3:
		if amt > 0 {
			return -amt
		}
		return amt
	case 4:
		if amt < 0 {
			return -amt
		}
		return amt
	default:
		return amt
	}
}
