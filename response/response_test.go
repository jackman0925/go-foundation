package response

import (
	"encoding/json"
	"testing"

	foundationErrors "github.com/jackman0925/go-foundation/errors"
)

func TestNewOKBuildsStandardResponse(t *testing.T) {
	resp := NewOK(map[string]string{"name": "demo"})

	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
	if resp.Message != "ok" {
		t.Fatalf("expected ok message, got %q", resp.Message)
	}
	if resp.Data == nil {
		t.Fatal("expected data")
	}
}

func TestNewFailUsesAppErrorCodeAndMessage(t *testing.T) {
	resp := NewFail(foundationErrors.New(40000, "参数错误"))

	if resp.Code != 40000 {
		t.Fatalf("expected code 40000, got %d", resp.Code)
	}
	if resp.Message != "参数错误" {
		t.Fatalf("expected message, got %q", resp.Message)
	}
}

func TestNewPageBuildsPagePayload(t *testing.T) {
	resp := NewPage([]int{1, 2}, 10, 2, 2)

	page, ok := resp.Data.(PageData)
	if !ok {
		t.Fatalf("expected PageData, got %T", resp.Data)
	}
	if page.Total != 10 || page.PageNo != 2 || page.PageSize != 2 {
		t.Fatalf("unexpected page payload: %+v", page)
	}
}

func TestPageDataMarshalUsesCamelCaseJSONNames(t *testing.T) {
	content, err := json.Marshal(PageData{
		List:     []int{1, 2},
		Total:    10,
		PageNo:   1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	expected := `{"list":[1,2],"total":10,"pageNo":1,"pageSize":20}`
	if string(content) != expected {
		t.Fatalf("expected %s, got %s", expected, content)
	}
}
