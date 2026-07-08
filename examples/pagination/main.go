package main

import "github.com/jackman0925/go-foundation/pagination"

func main() {
	page := pagination.Parse("2", "20")
	limit, offset := page.LimitOffset()
	_ = limit
	_ = offset
}
