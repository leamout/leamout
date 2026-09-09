package organizations

import "strconv"

const pageSize = 20

func pageFrom(raw string) int {
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 1
	}
	return page
}
