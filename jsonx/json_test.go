package jsonx

import "testing"

func TestMarshalToStringAndUnmarshalFromString(t *testing.T) {
	input := map[string]int{"count": 2}

	text, err := MarshalToString(input)
	if err != nil {
		t.Fatalf("MarshalToString returned error: %v", err)
	}

	var output map[string]int
	if err := UnmarshalFromString(text, &output); err != nil {
		t.Fatalf("UnmarshalFromString returned error: %v", err)
	}
	if output["count"] != 2 {
		t.Fatalf("unexpected output: %+v", output)
	}
}

func TestPrettyFormatsJSON(t *testing.T) {
	text, err := Pretty(map[string]int{"count": 2})
	if err != nil {
		t.Fatalf("Pretty returned error: %v", err)
	}
	if text == `{"count":2}` {
		t.Fatal("expected pretty formatted JSON")
	}
}

func TestMustToStringPanicsOnUnsupportedValue(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	_ = MustToString(make(chan int))
}
