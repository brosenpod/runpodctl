package format

import (
	"io"
	"reflect"
	"testing"

	"github.com/olekukonko/tablewriter"
)

func TestTableDefaults(t *testing.T) {
	table := tablewriter.NewWriter(io.Discard)

	// Flip a subset of settings away from their defaults so the test ensures
	// that TableDefaults actively applies the expected configuration.
	table.SetAutoWrapText(true)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_RIGHT)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetCenterSeparator("+")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	table.SetHeaderLine(true)
	table.SetBorder(true)
	table.SetTablePadding(" ")
	table.SetNoWhiteSpace(false)

	TableDefaults(table)

	assertBoolField(t, table, "autoWrap", false)
	assertBoolField(t, table, "autoFmt", true)
	assertIntField(t, table, "hAlign", tablewriter.ALIGN_LEFT)
	assertIntField(t, table, "align", tablewriter.ALIGN_LEFT)
	assertStringField(t, table, "pCenter", "")
	assertStringField(t, table, "pColumn", "")
	assertStringField(t, table, "pRow", "")
	assertBoolField(t, table, "hdrLine", false)
	assertBorderField(t, table, tablewriter.Border{Left: false, Right: false, Top: false, Bottom: false})
	assertStringField(t, table, "tablePadding", "\t")
	assertBoolField(t, table, "noWhiteSpace", true)
}

func tableField(t *testing.T, table *tablewriter.Table, name string) reflect.Value {
	t.Helper()

	v := reflect.ValueOf(table).Elem()
	field := v.FieldByName(name)
	if !field.IsValid() {
		t.Fatalf("tablewriter.Table field %q not found", name)
	}

	return field
}

func assertBoolField(t *testing.T, table *tablewriter.Table, name string, want bool) {
	t.Helper()

	if got := tableField(t, table, name).Bool(); got != want {
		t.Fatalf("tablewriter.Table.%s = %v, want %v", name, got, want)
	}
}

func assertIntField(t *testing.T, table *tablewriter.Table, name string, want int) {
	t.Helper()

	if got := int(tableField(t, table, name).Int()); got != want {
		t.Fatalf("tablewriter.Table.%s = %d, want %d", name, got, want)
	}
}

func assertStringField(t *testing.T, table *tablewriter.Table, name, want string) {
	t.Helper()

	if got := tableField(t, table, name).String(); got != want {
		t.Fatalf("tablewriter.Table.%s = %q, want %q", name, got, want)
	}
}

func assertBorderField(t *testing.T, table *tablewriter.Table, want tablewriter.Border) {
	t.Helper()

	border := tableField(t, table, "borders")
	checks := []struct {
		name string
		want bool
	}{
		{name: "Left", want: want.Left},
		{name: "Right", want: want.Right},
		{name: "Top", want: want.Top},
		{name: "Bottom", want: want.Bottom},
	}

	for _, check := range checks {
		if got := border.FieldByName(check.name).Bool(); got != check.want {
			t.Fatalf("tablewriter.Table.borders.%s = %v, want %v", check.name, got, check.want)
		}
	}
}
