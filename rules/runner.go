package rules

import (
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/lang"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/logger"
	"github.com/wata727/tflint/state"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Runner checks templates according rules.
// For variables interplation, it has Terraform eval context,
// and state. After checking, it accumulates results as issues.
type Runner struct {
	Issues issue.Issues

	ctx    terraform.BuiltinEvalContext
	state  state.TFState
	target *configs.Config
	config *config.Config
	logger *logger.Logger
}

// EvaluateExpr is a wrapper of terraform.BultinEvalContext.EvaluateExpr and gocty.FromCtyValue
// The `self` attribute is nil because the runner always handle input variables
// and terrfaorm attributes only. And also, variable type is `DynamicPseudoType`
// because it is hard to get the variable type defitinion.
func (r *Runner) EvaluateExpr(expr hcl.Expression, ret interface{}) error {
	val, diags := r.ctx.EvaluateExpr(expr, cty.DynamicPseudoType, nil)
	if diags.HasErrors() {
		return diags.Err()
	}
	err := gocty.FromCtyValue(val, ret)
	if err != nil {
		return err
	}
	return nil
}

func isEvaluable(expr hcl.Expression) bool {
	refs, diags := lang.ReferencesInExpr(expr)
	if diags.HasErrors() {
		panic(diags.Err())
	}
	for _, ref := range refs {
		switch ref.Subject.(type) {
		case addrs.InputVariable:
			// noop
		case addrs.TerraformAttr:
			// noop
		default:
			return false
		}
	}
	return true
}
