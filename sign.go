package main

func normalizeAmount(op int, amt float64) float64 {
	// PAYMENT (4) => positive, all others => negative
	if op == 4 {
		if amt < 0 {
			return -amt
		}
		return amt
	}
	if amt > 0 {
		return -amt
	}
	return amt
}
