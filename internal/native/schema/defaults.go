// Package schema provides runtime utilities for native generated resources,
// including static default value implementations.
package schema

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StaticBool returns a defaults.Bool that sets a static bool default.
func StaticBool(val bool) defaults.Bool {
	return staticBoolDefault{val: val}
}

type staticBoolDefault struct {
	val bool
}

func (d staticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %t", d.val)
}

func (d staticBoolDefault) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d staticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.val)
}

// StaticString returns a defaults.String that sets a static string default.
func StaticString(val string) defaults.String {
	return staticStringDefault{val: val}
}

type staticStringDefault struct {
	val string
}

func (d staticStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.val)
}

func (d staticStringDefault) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d staticStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.val)
}

// StaticInt64 returns a defaults.Int64 that sets a static int64 default.
func StaticInt64(val int64) defaults.Int64 {
	return staticInt64Default{val: val}
}

type staticInt64Default struct {
	val int64
}

func (d staticInt64Default) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %d", d.val)
}

func (d staticInt64Default) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d staticInt64Default) DefaultInt64(_ context.Context, _ defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(d.val)
}
