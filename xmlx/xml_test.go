package xmlx

import (
	"encoding/xml"
	"reflect"
	"strings"
	"testing"
)

type testUser struct {
	XMLName xml.Name `xml:"user"`
	ID      int64    `xml:"userId,attr"`
	Name    string   `xml:"userName"`
	Tags    []string `xml:"tags>tag"`
	Note    string   `xml:"note,omitempty"`
}

func TestStructRoundTrip(t *testing.T) {
	input := testUser{XMLName: xml.Name{Local: "user"}, ID: 1001, Name: "张三 & <开发>", Tags: []string{"开发", "管理"}}
	text, err := MarshalToString(input)
	if err != nil {
		t.Fatal(err)
	}
	want := `<user userId="1001"><userName>张三 &amp; &lt;开发&gt;</userName><tags><tag>开发</tag><tag>管理</tag></tags></user>`
	if text != want {
		t.Fatalf("got %s, want %s", text, want)
	}
	var output testUser
	if err := UnmarshalFromString(text, &output); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(input, output) {
		t.Fatalf("got %+v, want %+v", output, input)
	}
}

func TestPretty(t *testing.T) {
	value := struct {
		XMLName xml.Name `xml:"root"`
		Value   string   `xml:"value"`
	}{Value: "内容"}
	got, err := Pretty(value)
	if err != nil {
		t.Fatal(err)
	}
	if got != "<root>\n  <value>内容</value>\n</root>" {
		t.Fatalf("unexpected XML: %s", got)
	}
}

func TestMarshalErrors(t *testing.T) {
	for _, value := range []any{map[string]string{"key": "value"}, make(chan int), func() {}} {
		for _, encode := range []func(any) (string, error){MarshalToString, Pretty} {
			got, err := encode(value)
			if err == nil || got != "" {
				t.Fatalf("unsupported %T: got %q, %v", value, got, err)
			}
		}
	}
}

func TestUnmarshalErrors(t *testing.T) {
	for _, text := range []string{"", "<user>", "<other/>", "<user userId=\"invalid\"/>", "<user>&unknown;</user>"} {
		var user testUser
		if err := UnmarshalFromString(text, &user); err == nil {
			t.Fatalf("expected error for %q", text)
		}
	}
	for _, target := range []any{nil, testUser{}, (*testUser)(nil)} {
		if err := UnmarshalFromString("<user/>", target); err == nil {
			t.Fatalf("expected error for %T", target)
		}
	}
}

func TestNamespaceAndCharacterData(t *testing.T) {
	type message struct {
		XMLName xml.Name `xml:"urn:example message"`
		Text    string   `xml:",chardata"`
	}
	var got message
	if err := UnmarshalFromString(`<message xmlns="urn:example"><![CDATA[你好 <world>]]></message>`, &got); err != nil {
		t.Fatal(err)
	}
	if got.Text != "你好 <world>" || got.XMLName.Space != "urn:example" {
		t.Fatalf("unexpected result: %+v", got)
	}
	text, err := MarshalToString(got)
	if err != nil || !strings.Contains(text, "xmlns=\"urn:example\"") {
		t.Fatalf("got %q, %v", text, err)
	}
}
