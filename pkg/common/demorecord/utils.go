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
