package listquery

import (
	"net/url"
	"testing"
	"time"
)

func TestParserOptionalAndInvalidValues(t *testing.T) {
	for _, method := range []string{"text", "uuid", "bool", "time"} {
		t.Run(method, func(t *testing.T) {
			for _, query := range []string{"", "value=", "value=one&value=two", "value=invalid"} {
				values, _ := url.ParseQuery(query)
				p := Parser{Values: values}
				switch method {
				case "text":
					p.Text("value")
				case "uuid":
					p.UUID("value")
				case "bool":
					p.Bool("value")
				case "time":
					p.Time("value")
				}
				wantErr := query != "" && (method != "text" || query != "value=invalid")
				if (p.Err != nil) != wantErr {
					t.Fatalf("query %q error = %v", query, p.Err)
				}
			}
		})
	}
}

func TestTimestampNormalizesTimezone(t *testing.T) {
	p := Parser{Values: url.Values{"value": {"2026-01-01T02:00:00.123456+02:00"}}}
	value := p.Time("value")
	if p.Err != nil || value == nil || value.Location() != time.UTC || value.Format(time.RFC3339Nano) != "2026-01-01T00:00:00.123456Z" {
		t.Fatalf("time = %v, error = %v", value, p.Err)
	}
}

func TestRangeRejectsEqualReversedAndZeroBounds(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	before := from.Add(time.Hour)
	zero := time.Time{}
	for _, pair := range [][2]*time.Time{{&from, &from}, {&before, &from}, {&zero, nil}, {nil, &zero}} {
		if Range(pair[0], pair[1]) == nil {
			t.Fatal("expected invalid range")
		}
	}
	for _, pair := range [][2]*time.Time{{nil, nil}, {&from, nil}, {nil, &before}, {&from, &before}} {
		if err := Range(pair[0], pair[1]); err != nil {
			t.Fatal(err)
		}
	}
}
