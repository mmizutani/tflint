package rules

import (
	"fmt"
	"path/filepath"

	instances "github.com/cristim/ec2-instances-info"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// CheckAwsInstanceInvalidType checks whether "aws_instance"
// has invalid instance type.
// TODO: DRY
func (r *Runner) CheckAwsInstanceInvalidType() {
	instanceTypes := map[string]bool{}
	data, err := instances.Data()
	if err != nil {
		// TODO
		panic(err)
	}

	for _, i := range *data {
		instanceTypes[i.InstanceType] = true
	}

	for _, resource := range r.target.Module.ManagedResources {
		if resource.Type != "aws_instance" {
			continue
		}

		body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{
					Name: "instance_type",
				},
			},
		})
		if diags.HasErrors() {
			// TODO
			panic(diags)
		}

		if attribute, ok := body.Attributes["instance_type"]; ok && isEvaluable(attribute.Expr) {
			val, diags := r.ctx.EvaluateExpr(attribute.Expr, cty.DynamicPseudoType, nil)
			if diags.HasErrors() {
				// TODO
				panic(diags.Err())
			}

			var instanceType string
			err = gocty.FromCtyValue(val, &instanceType)
			if err != nil {
				// TODO
				panic(err)
			}

			if !instanceTypes[instanceType] {
				r.Issues = append(r.Issues, &issue.Issue{
					Detector: "aws_instance_invalid_type",
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid instance type.", instanceType),
					Line:     attribute.Range.Start.Line,
					File:     filepath.Base(attribute.Range.Filename),
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_invalid_type.md",
				})
			}
		}
	}
}
