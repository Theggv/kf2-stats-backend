package demorecord

func readByte(data []byte, pos int) byte {
	return data[pos]
}

func readInt(data []byte, pos int) int {
	return (int(data[pos]) << 24) + (int(data[pos+1]) << 16) + (int(data[pos+2]) << 8) + int(data[pos+3])
}

func readString(data []byte, pos, size int) string {
	return string(data[pos : pos+size])
}

func lerp(start, end, t float64) float64 {
	return start*(1-t) + end*t
}

func standardize(value, mean, stddev float64) float64 {
	return (value - mean) / stddev
}

func mean(value []float64) float64 {
	sum := 0.0

	for i := range value {
		sum += value[i]
	}

	return sum / float64(len(value))
}
