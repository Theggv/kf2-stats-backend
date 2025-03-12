package demorecord

import "cmp"

func findFirstGreaterIndex[K cmp.Ordered, V any](data []V, key func(V) K, needle K) int {
	if len(data) == 0 {
		return -1
	}

	lo := 0
	hi := len(data) - 1
	mid := 0

	for lo <= hi {
		mid = (lo + hi) >> 1
		value := key(data[mid])

		if value == needle {
			for i := mid - 1; i >= lo; i-- {
				if key(data[i]) == needle {
					mid = i
				} else {
					break
				}
			}

			return mid
		} else if value < needle {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}

	if lo >= len(data) {
		return -1
	}

	return lo
}

func findLastLowerIndex[K cmp.Ordered, V any](data []V, key func(V) K, needle K) int {
	if len(data) == 0 {
		return -1
	}

	lo := 0
	hi := len(data) - 1
	mid := 0

	for lo <= hi {
		mid = (lo + hi) >> 1
		value := key(data[mid])

		if value == needle {
			for i := mid + 1; i <= hi; i++ {
				if key(data[i]) == needle {
					mid = i
				} else {
					break
				}
			}

			return mid
		} else if value < needle {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}

	return hi
}

func findFirstGreater[K cmp.Ordered, V any](data []V, key func(V) K, needle K) *V {
	idx := findFirstGreaterIndex(data, key, needle)

	if idx < 0 {
		return nil
	}

	return &data[idx]
}

func findLastLower[K cmp.Ordered, V any](data []V, key func(V) K, needle K) *V {
	idx := findLastLowerIndex(data, key, needle)

	if idx < 0 {
		return nil
	}

	return &data[idx]
}

func filterByRange[K cmp.Ordered, V any](data []V, key func(item V) K, from, to K) []V {
	startIdx := findFirstGreaterIndex(data, key, from)
	endIdx := findLastLowerIndex(data, key, to)

	if startIdx < 0 || endIdx < 0 {
		return []V{}
	}

	return data[startIdx : endIdx+1]
}

func filter[K cmp.Ordered, V any](data []V, key func(item V) K, value K) []V {
	res := []V{}

	for i := range data {
		item := data[i]

		if key(item) == value {
			res = append(res, item)
		}
	}

	return res
}
