// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/opentofu/opentofu/internal/configs/configschema"
	"github.com/zclconf/go-cty/cty"
)

func TestInputForSchema_contextCanceled(t *testing.T) {
	view, _ := testView(t)
	m := &Meta{
		View:  view,
		input: true,
	}

	schema := &configschema.Block{
		Attributes: map[string]*configschema.Attribute{
			"address": {
				Type:        cty.String,
				Required:    true,
				Description: "The HTTP backend address.",
			},
		},
	}

	given := cty.ObjectVal(map[string]cty.Value{
		"address": cty.NullVal(cty.String),
	})

	r, w := io.Pipe()
	defaultInputReader = r
	defaultInputWriter = io.Discard
	t.Cleanup(func() {
		defaultInputReader = nil
		defaultInputWriter = nil
		_ = w.Close()
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := m.inputForSchema(ctx, given, schema, view)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}

func TestInputForSchema_success(t *testing.T) {
	view, _ := testView(t)
	m := &Meta{
		View:  view,
		input: true,
	}

	schema := &configschema.Block{
		Attributes: map[string]*configschema.Attribute{
			"address": {
				Type:        cty.String,
				Required:    true,
				Description: "The HTTP backend address.",
			},
		},
	}

	given := cty.ObjectVal(map[string]cty.Value{
		"address": cty.NullVal(cty.String),
	})

	testInputResponseMap = map[string]string{
		"address": "http://localhost:8080",
	}
	t.Cleanup(func() {
		testInputResponseMap = nil
	})

	val, err := m.inputForSchema(t.Context(), given, schema, view)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := cty.ObjectVal(map[string]cty.Value{
		"address": cty.StringVal("http://localhost:8080"),
	})
	if !val.RawEquals(expected) {
		t.Fatalf("got %#v, want %#v", val, expected)
	}
}
