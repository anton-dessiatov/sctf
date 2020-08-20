package terra

import (
	"github.com/hashicorp/terraform/backend/local"
	"github.com/hashicorp/terraform/plans"
	"github.com/hashicorp/terraform/terraform"
)

func (t *Terra) renderPlan(plan *plans.Plan, schemas *terraform.Schemas) {
	local.RenderPlan(plan, nil, nil, schemas, t.ui, &t.colors)
}
