package convert

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringValue(t *testing.T) {
	if got := StringValue(nil); !got.IsNull() {
		t.Fatalf("StringValue(nil) = %v, want null", got)
	}

	value := "hello"
	if got := StringValue(&value); got.ValueString() != "hello" {
		t.Fatalf("StringValue() = %v", got)
	}
}

func TestStringValueOr(t *testing.T) {
	if got := StringValueOr(nil, "fallback"); got.ValueString() != "fallback" {
		t.Fatalf("StringValueOr() = %v", got)
	}
}

func TestStringPointer(t *testing.T) {
	if got := StringPointer(types.StringNull()); got != nil {
		t.Fatalf("StringPointer(null) = %v, want nil", got)
	}
	if got := StringPointer(types.StringUnknown()); got != nil {
		t.Fatalf("StringPointer(unknown) = %v, want nil", got)
	}

	got := StringPointer(types.StringValue("hello"))
	if got == nil || *got != "hello" {
		t.Fatalf("StringPointer() = %v", got)
	}
}

func TestBoolValue(t *testing.T) {
	if got := BoolValue(nil); !got.IsNull() {
		t.Fatalf("BoolValue(nil) = %v, want null", got)
	}
	if got := BoolValueOr(nil, false); got.ValueBool() {
		t.Fatal("BoolValueOr(nil, false) = true")
	}

	value := true
	if got := BoolValue(&value); !got.ValueBool() {
		t.Fatal("BoolValue(true) = false")
	}
}

func TestBoolPointer(t *testing.T) {
	if got := BoolPointer(types.BoolNull()); got != nil {
		t.Fatalf("BoolPointer(null) = %v, want nil", got)
	}

	got := BoolPointer(types.BoolValue(true))
	if got == nil || !*got {
		t.Fatalf("BoolPointer() = %v", got)
	}
}

func TestFloatAndIntConversion(t *testing.T) {
	version := 0.3
	if got := Float64Value(&version); got.ValueFloat64() != 0.3 {
		t.Fatalf("Float64Value() = %v", got)
	}
	if got := Float64Value(nil); !got.IsNull() {
		t.Fatalf("Float64Value(nil) = %v, want null", got)
	}

	updatedAt := int64(1_700_000_000_000)
	if got := Int64Value(&updatedAt); got.ValueInt64() != updatedAt {
		t.Fatalf("Int64Value() = %v", got)
	}
	if got := Int64Value(nil); !got.IsNull() {
		t.Fatalf("Int64Value(nil) = %v, want null", got)
	}

	count := int32(7)
	if got := Int32Value(&count); got.ValueInt64() != 7 {
		t.Fatalf("Int32Value() = %v", got)
	}
}

func TestTimestampValue(t *testing.T) {
	if got := TimestampValue(nil); !got.IsNull() {
		t.Fatalf("TimestampValue(nil) = %v, want null", got)
	}

	zero := int64(0)
	if got := TimestampValue(&zero); !got.IsNull() {
		t.Fatalf("TimestampValue(0) = %v, want null", got)
	}

	millis := time.Date(2026, 10, 25, 12, 30, 0, 0, time.UTC).UnixMilli()
	got := TimestampValue(&millis)
	if got.ValueString() != "2026-10-25T12:30:00Z" {
		t.Fatalf("TimestampValue() = %q", got.ValueString())
	}
}

func TestStringListRoundTrip(t *testing.T) {
	ctx := context.Background()

	list, diags := StringListValue(ctx, []string{"a", "b"})
	if diags.HasError() {
		t.Fatalf("StringListValue() diags = %v", diags)
	}

	values, diags := StringList(ctx, list)
	if diags.HasError() {
		t.Fatalf("StringList() diags = %v", diags)
	}
	if len(values) != 2 || values[0] != "a" || values[1] != "b" {
		t.Fatalf("values = %v", values)
	}

	nullList, diags := StringListValue(ctx, nil)
	if diags.HasError() {
		t.Fatalf("StringListValue(nil) diags = %v", diags)
	}
	if !nullList.IsNull() {
		t.Fatalf("StringListValue(nil) = %v, want null", nullList)
	}

	values, diags = StringList(ctx, types.ListNull(types.StringType))
	if diags.HasError() || values != nil {
		t.Fatalf("StringList(null) = %v, %v", values, diags)
	}
}
