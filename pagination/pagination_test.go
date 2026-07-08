package pagination

import (
	"encoding/json"
	"testing"
)

func TestParseUsesDefaultsForInvalidInput(t *testing.T) {
	page := Parse("0", "bad")

	if page.PageNo != 1 {
		t.Fatalf("expected default pageNo 1, got %d", page.PageNo)
	}
	if page.PageSize != 20 {
		t.Fatalf("expected default page size 20, got %d", page.PageSize)
	}
}

func TestParseCapsPageSizeAndCalculatesLimitOffset(t *testing.T) {
	page := Parse("3", "1000")
	limit, offset := page.LimitOffset()

	if page.PageSize != 100 {
		t.Fatalf("expected max page size 100, got %d", page.PageSize)
	}
	if limit != 100 || offset != 200 {
		t.Fatalf("expected limit=100 offset=200, got limit=%d offset=%d", limit, offset)
	}
}

func TestPageMarshalUsesCamelCaseJSONNames(t *testing.T) {
	content, err := json.Marshal(Page{PageNo: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	if string(content) != `{"pageNo":1,"pageSize":20}` {
		t.Fatalf("unexpected JSON: %s", content)
	}
}
