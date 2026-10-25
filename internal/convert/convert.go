// Package convert maps between OpenMetadata SDK types and Terraform framework
// types. Entity packages keep their own nested-shape helpers here only when the
// mapping is shared by more than one entity.
package convert

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringValue maps a nullable API string to a Terraform string.
func StringValue(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

// StringValueOr maps a nullable API string, falling back when it is absent.
func StringValueOr(value *string, fallback string) types.String {
	if value == nil {
		return types.StringValue(fallback)
	}
	return types.StringValue(*value)
}

// StringPointer maps a Terraform string to an API pointer. Null and unknown
// values become nil so the field is omitted from the request.
func StringPointer(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	s := value.ValueString()
	return &s
}

// BoolValue maps a nullable API bool to a Terraform bool.
func BoolValue(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}

// BoolValueOr maps a nullable API bool, falling back when it is absent. Use it
// for fields whose Terraform schema carries a default so state stays stable.
func BoolValueOr(value *bool, fallback bool) types.Bool {
	if value == nil {
		return types.BoolValue(fallback)
	}
	return types.BoolValue(*value)
}

// BoolPointer maps a Terraform bool to an API pointer.
func BoolPointer(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	b := value.ValueBool()
	return &b
}

// Float64Value maps a nullable API float to a Terraform float.
func Float64Value(value *float64) types.Float64 {
	if value == nil {
		return types.Float64Null()
	}
	return types.Float64Value(*value)
}

// Int64Value maps a nullable API int64 to a Terraform number.
func Int64Value(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}

// Int32Value maps a nullable API int32 to a Terraform number.
func Int32Value(value *int32) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*value))
}

// TimestampValue renders OpenMetadata's epoch-millisecond timestamp as
// RFC 3339 in UTC. Missing and non-positive values become null.
func TimestampValue(epochMillis *int64) types.String {
	if epochMillis == nil || *epochMillis <= 0 {
		return types.StringNull()
	}
	return types.StringValue(time.UnixMilli(*epochMillis).UTC().Format(time.RFC3339))
}

// StringList maps a Terraform list to a slice. A null list yields nil so the
// caller can omit the field from an API request.
func StringList(ctx context.Context, value types.List) ([]string, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}
	var out []string
	diags := value.ElementsAs(ctx, &out, false)
	return out, diags
}

// StringListValue maps a slice to a Terraform list. A nil slice becomes null.
func StringListValue(ctx context.Context, values []string) (types.List, diag.Diagnostics) {
	if values == nil {
		return types.ListNull(types.StringType), nil
	}
	return types.ListValueFrom(ctx, types.StringType, values)
}
