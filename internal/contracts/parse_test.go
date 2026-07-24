package contracts

import (
	"reflect"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	in := "# comment\nFOO=bar\n\n  BAZ = qux \nMALFORMED\n"
	got := ParseEnvFile(in)
	want := map[string]string{"FOO": "bar", "BAZ": "qux"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMergeMapsOverrideWins(t *testing.T) {
	base := map[string]string{"A": "1", "B": "2"}
	over := map[string]string{"B": "9", "C": "3"}
	got := MergeMaps(base, over)
	want := map[string]string{"A": "1", "B": "9", "C": "3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseSecretRefsJSONPatchShape(t *testing.T) {
	in := `
- op: add
  path: /spec/data/-
  value:
    secretKey: DB_PASSWORD
    remoteRef:
      key: prod-db-password
`
	got := ParseSecretRefs(in)
	if got["DB_PASSWORD"] != "prod-db-password" {
		t.Fatalf("got %v", got)
	}
}

func TestParseSecretRefsPlainShape(t *testing.T) {
	in := "spec:\n  data:\n    - secretKey: \"API_KEY\"\n      remoteRef:\n        key: 'prod-api-key'\n"
	got := ParseSecretRefs(in)
	if got["API_KEY"] != "prod-api-key" {
		t.Fatalf("got %v", got)
	}
}

func TestParseSecretRefsMultiKeyList(t *testing.T) {
	in := "spec:\n  data:\n    - secretKey: FOO_KEY\n      remoteRef:\n        key: foo\n    - secretKey: BAR_KEY\n      remoteRef:\n        key: bar\n"
	got := ParseSecretRefs(in)
	want := map[string]string{"FOO_KEY": "foo", "BAR_KEY": "bar"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseSecretRefsMalformedNoPanic(t *testing.T) {
	in := "not: [valid: yaml: -"
	got := ParseSecretRefs(in)
	if len(got) != 0 {
		t.Fatalf("got %v, want empty map", got)
	}
}
