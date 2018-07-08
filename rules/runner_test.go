package rules

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

// TODO: Handle type mismatch, interpolation errors, terraform attributes, block, etc.

func Test_EvaluateExpr_string(t *testing.T) {
	cases := []struct {
		Name      string
		Key       string
		Variables map[string]map[string]cty.Value
		Expected  string
	}{
		{
			Name:      "literal",
			Key:       "literal",
			Variables: map[string]map[string]cty.Value{},
			Expected:  "literal_val",
		},
		{
			Name: "string interpolation",
			Key:  "string",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"string_var": cty.StringVal("string_val"),
				},
			},
			Expected: "string_val",
		},
		{
			Name: "new style interpolation",
			Key:  "new_string",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"string_var": cty.StringVal("string_val"),
				},
			},
			Expected: "string_val",
		},
		{
			Name: "list element",
			Key:  "list_element",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"list_var": cty.TupleVal([]cty.Value{cty.StringVal("one"), cty.StringVal("two")}),
				},
			},
			Expected: "one",
		},
		{
			Name: "map element",
			Key:  "map_element",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"map_var": cty.ObjectVal(map[string]cty.Value{"one": cty.StringVal("one"), "two": cty.StringVal("two")}),
				},
			},
			Expected: "one",
		},
		{
			Name:      "conditional",
			Key:       "conditional",
			Variables: map[string]map[string]cty.Value{},
			Expected:  "production",
		},
		{
			Name:      "bulit-in function",
			Key:       "function",
			Variables: map[string]map[string]cty.Value{},
			Expected:  "acbd18db4cc2f85cedef654fccc4a4d8",
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
		}

		var ret string
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatal(err)
		}

		if tc.Expected != ret {
			t.Fatalf("Expected value is not matched:\n Test: %s\n Expected: %s\n Actual: %s\n", tc.Name, tc.Expected, ret)
		}
	}
}

func Test_EvaluateExpr_integer(t *testing.T) {
	cases := []struct {
		Name      string
		Key       string
		Variables map[string]map[string]cty.Value
		Expected  int
	}{
		{
			Name: "integer interpolation",
			Key:  "integer",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"integer_var": cty.NumberIntVal(3),
				},
			},
			Expected: 3,
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
		}

		var ret int
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatal(err)
		}

		if tc.Expected != ret {
			t.Fatalf("Expected value is not matched:\n Test: %s\n Expected: %d\n Actual: %d\n", tc.Name, tc.Expected, ret)
		}
	}
}

func Test_EvaluateExpr_list(t *testing.T) {
	// TODO: How to handle a list when element count is unknown?
	type list struct {
		Elem1 string
		Elem2 string
		Elem3 string
	}

	cases := []struct {
		Name      string
		Key       string
		Variables map[string]map[string]cty.Value
		Expected  list
	}{
		{
			Name:      "list literal",
			Key:       "list",
			Variables: map[string]map[string]cty.Value{},
			Expected:  list{Elem1: "one", Elem2: "two", Elem3: "three"},
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
		}

		ret := list{}
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(tc.Expected, ret) {
			t.Fatalf("Expected value is not matched:\n Test: %s\n Diff: %s\n", tc.Name, cmp.Diff(tc.Expected, ret))
		}
	}
}

func Test_EvaluateExpr_map(t *testing.T) {
	type mapObject struct {
		One int `cty:"one"`
		Two int `cty:"two"`
	}

	cases := []struct {
		Name      string
		Key       string
		Variables map[string]map[string]cty.Value
		Expected  mapObject
	}{
		{
			Name:      "map literal",
			Key:       "map",
			Variables: map[string]map[string]cty.Value{},
			Expected:  mapObject{One: 1, Two: 2},
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
		}

		ret := mapObject{}
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(tc.Expected, ret) {
			t.Fatalf("Expected value is not matched:\n Test: %s\n Diff: %s\n", tc.Name, cmp.Diff(tc.Expected, ret))
		}
	}
}

func loadConfigHelper() (*configs.Config, error) {
	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		return nil, err
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	mod, diags := loader.Parser().LoadConfigDir(dir + "/test-fixtures/runner")
	if diags.HasErrors() {
		return nil, diags
	}
	cfg, diags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
	if diags.HasErrors() {
		return nil, diags
	}

	return cfg, nil
}

func extractAttributeHelper(key string, cfg *configs.Config) (*hcl.Attribute, error) {
	resource := cfg.Module.ManagedResources["null_resource.test"]
	if resource == nil {
		return nil, errors.New("Expected resource is not found")
	}
	body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: key,
			},
		},
	})
	if diags.HasErrors() {
		return nil, diags
	}
	attribute := body.Attributes[key]
	if attribute == nil {
		return nil, fmt.Errorf("Expected attribute is not found: %s", key)
	}
	return attribute, nil
}
