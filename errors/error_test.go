package errors

import (
	stderrors "errors"
	"testing"
)

func TestNewCreatesAppError(t *testing.T) {
	err := New(40000, "参数错误")

	var appErr *AppError
	if !stderrors.As(err, &appErr) {
		t.Fatal("expected AppError")
	}
	if appErr.Code != 40000 {
		t.Fatalf("expected code 40000, got %d", appErr.Code)
	}
	if appErr.Message != "参数错误" {
		t.Fatalf("expected message, got %q", appErr.Message)
	}
	if err.Error() != "参数错误" {
		t.Fatalf("unexpected error string: %q", err.Error())
	}
}

func TestWrapPreservesCause(t *testing.T) {
	cause := stderrors.New("database unavailable")
	err := Wrap(50000, "查询失败", cause)

	if !stderrors.Is(err, cause) {
		t.Fatal("expected wrapped cause to be discoverable")
	}
	if CodeOf(err) != 50000 {
		t.Fatalf("expected code 50000, got %d", CodeOf(err))
	}
	if MessageOf(err) != "查询失败" {
		t.Fatalf("expected wrapped message, got %q", MessageOf(err))
	}
}
