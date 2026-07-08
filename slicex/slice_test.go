package slicex

import (
	"reflect"
	"testing"
)

func TestContains(t *testing.T) {
	if !Contains([]int{1, 2, 3}, 2) {
		t.Fatal("expected slice to contain value")
	}
	if Contains([]int{1, 2, 3}, 4) {
		t.Fatal("expected slice to not contain value")
	}
}

func TestUniquePreservesFirstSeenOrder(t *testing.T) {
	got := Unique([]int64{3, 1, 3, 2, 1})
	expected := []int64{3, 1, 2}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestReverse(t *testing.T) {
	values := []string{"a", "b", "c"}
	Reverse(values)
	expected := []string{"c", "b", "a"}
	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("expected %v, got %v", expected, values)
	}
}
