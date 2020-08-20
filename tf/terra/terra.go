package terra

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statemgr"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/terraform/tfdiags"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/colorstring"
	"github.com/zclconf/go-cty/cty"
)

// Terra is the wrapper around Terraform capable of doing simple operations
type Terra struct {
	db          *gorm.DB
	credentials Credentials

	ui     cli.Ui
	colors colorstring.Colorize
}

func NewTerra(db *gorm.DB, credentials Credentials) *Terra {
	return &Terra{
		db:          db,
		credentials: credentials,
		ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}
}

func (t *Terra) context(opts terraform.ContextOpts, stack Stack, si *StackIdentity) (*terraform.Context, statemgr.Full, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics

	m, err, hclDiags := stack.Config.ToModule()
	if err != nil {
		diags = diags.Append(err)
		return nil, nil, diags
	}

	diags = diags.Append(hclDiags)
	if hclDiags.HasErrors() {
		return nil, nil, diags
	}

	providerConfigs, pcDiags := t.providerConfigs(stack.AWS, stack.GCP)
	diags = diags.Append(pcDiags)
	if pcDiags.HasErrors() {
		return nil, nil, diags
	}

	m.ProviderConfigs = providerConfigs

	opts.Config = &configs.Config{
		Module: m,
	}
	opts.Providers = t.providers()

	var st statemgr.Full
	if si != nil {
		st, err = t.NewStateMgr(*si)
		if err != nil {
			diags = diags.Append(fmt.Errorf("t.NewStateMgr: %w", err))
			return nil, nil, diags
		}
		if err := st.RefreshState(); err != nil {
			diags = diags.Append(fmt.Errorf("st.RefreshState: %w", err))
			return nil, nil, diags
		}

		opts.State = st.State()
	}

	result, newCtxDiags := terraform.NewContext(&opts)
	diags = diags.Append(newCtxDiags)

	return result, st, diags
}

func (t *Terra) Apply(si StackIdentity, s Stack, destroy bool) tfdiags.Diagnostics {
	tfCtx, st, diags := t.context(terraform.ContextOpts{Destroy: destroy}, s, &si)
	if diags.HasErrors() {
		return diags
	}

	plan, planDiags := tfCtx.Plan()
	diags = diags.Append(planDiags)
	if planDiags.HasErrors() {
		return diags
	}

	t.renderPlan(plan, tfCtx.Schemas())

	_, applyDiags := tfCtx.Apply()
	diags = diags.Append(applyDiags)
	// We need to persist the state even if apply failed
	applyState := tfCtx.State()

	err := statemgr.WriteAndPersist(st, applyState)
	if err != nil {
		diags = diags.Append(err)
		return diags
	}

	if applyDiags.HasErrors() {
		return diags
	}

	return diags
}

// Evaluate returns the resource instance for a given address at a given state. We also need
// a stack definition used to produce that state to get schemas
func (t *Terra) Evaluate(s Stack, st *states.State, addr addrs.AbsResourceInstance) (cty.Value, error) {
	tfCtx, _, diags := t.context(terraform.ContextOpts{Destroy: false}, s, nil)
	if diags.HasErrors() {
		return cty.Value{}, diags.Err()
	}

	schemas := tfCtx.Schemas()

	res := st.Resource(addr.ContainingResource())
	if res == nil {
		return cty.Value{}, fmt.Errorf("failed to find resource for %v", addr)
	}

	ri := res.Instance(addr.Resource.Key)
	if ri == nil {
		return cty.Value{}, fmt.Errorf("failed to find instance for %v", addr)
	}

	provider := res.ProviderConfig.Provider
	if _, exists := schemas.Providers[provider]; !exists {
		return cty.Value{}, fmt.Errorf("failed to locate provider for %v", addr)
	}

	schema, _ := schemas.ResourceTypeConfig(provider, addr.Resource.Resource.Mode, addr.Resource.Resource.Type)
	if schema == nil {
		return cty.Value{}, fmt.Errorf("schema missing for %v", addr)
	}

	val, err := ri.Current.Decode(schema.ImpliedType())
	if err != nil {
		return cty.Value{}, fmt.Errorf("instance.Decode: %w", err)
	}

	return val.Value, nil
}
