package rules

import (
	"os"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/issue"
	"github.com/zclconf/go-cty/cty"
)

func Test_CheckAwsInstanceInvalidType(t *testing.T) {
	cases := []struct {
		Name      string
		Dir       string
		Variables map[string]map[string]cty.Value
		Issues    issue.Issues
	}{
		{
			Name:      "literal",
			Dir:       "literal",
			Variables: map[string]map[string]cty.Value{},
			Issues: []*issue.Issue{
				{
					Detector: "aws_instance_invalid_type",
					Type:     issue.ERROR,
					Message:  "\"t1.2xlarge\" is invalid instance type.",
					Line:     2,
					File:     "instances.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_invalid_type.md",
				},
			},
		},
		{
			Name:      "missing_key",
			Dir:       "missing_key",
			Variables: map[string]map[string]cty.Value{},
			Issues:    []*issue.Issue{},
		},
		{
			Name: "eval_string",
			Dir:  "variable",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"instance_type": cty.StringVal("t1.2xlarge"),
				},
			},
			Issues: []*issue.Issue{
				{
					Detector: "aws_instance_invalid_type",
					Type:     issue.ERROR,
					Message:  "\"t1.2xlarge\" is invalid instance type.",
					Line:     4,
					File:     "instances.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_invalid_type.md",
				},
			},
		},
	}

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		mod, diags := loader.Parser().LoadConfigDir(dir + "/test-fixtures/aws_instance_invalid_type/" + tc.Dir)
		if diags.HasErrors() {
			panic(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			panic(tfdiags)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
			target: cfg,
			Issues: []*issue.Issue{},
		}

		runner.CheckAwsInstanceInvalidType()

		if !cmp.Equal(tc.Issues, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Issues, runner.Issues))
		}
	}
}
