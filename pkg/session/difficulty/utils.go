package difficulty

type pair struct {
	From, To float64
}

func npInterp(x float64, xp, fp pair) float64 {
	if x < xp.From {
		return fp.From
	}
	if x > xp.To {
		return fp.To
	}

	t := (x - xp.From) / (xp.To - xp.From)

	return fp.From*(1-t) + fp.To*t
}
