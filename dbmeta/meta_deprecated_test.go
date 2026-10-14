package dbmeta

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func loadTestMappings(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	mappingFile := filepath.Join(filepath.Dir(filename), "..", "template", "mapping.json")
	if err := LoadMappings(mappingFile, false); err != nil {
		t.Fatalf("LoadMappings: %v", err)
	}
}

func TestIsDeprecatedColumnComment(t *testing.T) {
	tests := []struct {
		comment  string
		expected bool
	}{
		{comment: "", expected: false},
		{comment: "go_type:\"github.com/foo:Bar\"", expected: false},
		{comment: "DEPRECATED: do not read or write", expected: true},
		{comment: "  DEPRECATED: spaced prefix", expected: true},
		{comment: "deprecated: legacy column", expected: false},
		{comment: "Deprecated at timestamp", expected: false},
		{comment: "DEPRECATED flag for the row", expected: false},
		{comment: "DEPRECATED", expected: false},
		{comment: "see DEPRECATED: elsewhere", expected: false},
	}
	for _, testCase := range tests {
		// When
		got := isDeprecatedColumnComment(testCase.comment)

		// Then
		if got != testCase.expected {
			t.Fatalf("isDeprecatedColumnComment(%q) = %v, want %v", testCase.comment, got, testCase.expected)
		}
	}
}

func TestGenerateFieldsTypesSkipsDeprecatedColumns(t *testing.T) {
	loadTestMappings(t)

	// Given
	dbMeta := &dbTableMeta{
		tableSchemaAndName: TableSchemaAndName{TableSchema: "config", TableName: "config_versions"},
		columns: []*columnMeta{
			{
				index: 0, name: "version_id", databaseTypeName: "text",
				colDDL: "TEXT", columnType: "TEXT", nullable: false, isPrimaryKey: true,
			},
			{
				index: 1, name: "saved_response", databaseTypeName: "jsonb",
				colDDL: "JSONB", columnType: "JSONB", nullable: true,
				comment: "DEPRECATED: do not read or write",
			},
		},
	}
	conf := &Config{FieldNamingTemplate: "{{FmtFieldName .}}", UseGureguTypes: false}

	// When
	fields, _, err := conf.GenerateFieldsTypes(dbMeta)

	// Then
	if err != nil {
		t.Fatalf("GenerateFieldsTypes: %v", err)
	}
	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}
	if fields[0].GoFieldName != "VersionID" {
		t.Fatalf("expected VersionID field, got %q", fields[0].GoFieldName)
	}
	if strings.Contains(fields[0].Code, "saved_response") {
		t.Fatalf("deprecated column should not appear in generated field code: %s", fields[0].Code)
	}
}

func TestGenerateFieldsTypesRejectsDeprecatedPrimaryKey(t *testing.T) {
	// Given
	dbMeta := &dbTableMeta{
		tableSchemaAndName: TableSchemaAndName{TableName: "bad_table"},
		columns: []*columnMeta{
			{
				index: 0, name: "id", databaseTypeName: "text",
				colDDL: "TEXT", columnType: "TEXT", nullable: false, isPrimaryKey: true,
				comment: "DEPRECATED: cannot drop yet",
			},
		},
	}
	conf := &Config{FieldNamingTemplate: "{{FmtFieldName .}}"}

	// When
	_, _, err := conf.GenerateFieldsTypes(dbMeta)

	// Then
	if err == nil {
		t.Fatal("expected error deprecating primary key column")
	}
}

func TestGenerateFieldsTypesDeprecatedNotNullColumn(t *testing.T) {
	loadTestMappings(t)

	tests := []struct {
		name            string
		defaultVal      string
		isAutoIncrement bool
		expectErr       bool
	}{
		{name: "no default", expectErr: true},
		{name: "with default", defaultVal: "{}"},
		{name: "identity", isAutoIncrement: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Given
			dbMeta := &dbTableMeta{
				tableSchemaAndName: TableSchemaAndName{TableName: "config_versions"},
				columns: []*columnMeta{
					{
						index: 0, name: "version_id", databaseTypeName: "text",
						colDDL: "TEXT", columnType: "TEXT", nullable: false, isPrimaryKey: true,
					},
					{
						index: 1, name: "saved_response", databaseTypeName: "jsonb",
						colDDL: "JSONB", columnType: "JSONB", nullable: false,
						defaultVal: testCase.defaultVal, isAutoIncrement: testCase.isAutoIncrement,
						comment: "DEPRECATED: do not read or write",
					},
				},
			}
			conf := &Config{FieldNamingTemplate: "{{FmtFieldName .}}"}

			// When
			fields, _, err := conf.GenerateFieldsTypes(dbMeta)

			// Then
			if testCase.expectErr {
				if err == nil {
					t.Fatal("expected error deprecating NOT NULL column without a default")
				}
				return
			}
			if err != nil {
				t.Fatalf("GenerateFieldsTypes: %v", err)
			}
			if len(fields) != 1 {
				t.Fatalf("expected 1 field, got %d", len(fields))
			}
		})
	}
}
