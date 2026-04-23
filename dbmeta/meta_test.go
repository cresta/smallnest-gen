package dbmeta

import "testing"

func TestPrimaryKeyParseAsFromComment(t *testing.T) {
	t.Parallel()
	if got := primaryKeyParseAsFromComment(""); got != "" {
		t.Fatalf("empty: got %q", got)
	}
	if got := primaryKeyParseAsFromComment(`primary_key_parse_as:"int32"`); got != "int32" {
		t.Fatalf("tag only: got %q", got)
	}
	if got := primaryKeyParseAsFromComment(`foo:"bar" primary_key_parse_as:"int64" baz:"qux"`); got != "int64" {
		t.Fatalf("embedded tag: got %q", got)
	}
}
