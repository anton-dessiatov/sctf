package cluster

import (
	"fmt"

	"github.com/anton-dessiatov/sctf/tf/dal"
	"github.com/anton-dessiatov/sctf/tf/model"
	"github.com/anton-dessiatov/sctf/tf/terra"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/states/statemgr"
	"github.com/jinzhu/gorm"
	"github.com/zclconf/go-cty/cty"
)

func StateByClusterID(db *gorm.DB, terra *terra.Terra, clusterID int) (model.ClusterState, error) {
	cluster, err := dal.ClusterByID(db, clusterID)
	if err != nil {
		return model.ClusterState{}, fmt.Errorf("dal.ClusterByID: %w", err)
	}

	stack, err := Stack(model.ClusterIdentity(clusterID), cluster.Template)
	if err != nil {
		return model.ClusterState{}, fmt.Errorf("Stack: %w", err)
	}

	sm, err := terra.NewStateMgr(StackIdentity(clusterID))
	if err != nil {
		return model.ClusterState{}, fmt.Errorf("terra.NewStateMgr: %w", err)
	}

	sb := StateBuilder{
		ClusterTemplate: cluster.Template,
		Stack:           stack,
		StateMgr:        sm,
		Terra:           terra,
	}

	result, err := sb.Build()
	if err != nil {
		return result, fmt.Errorf("sb.Build: %w", err)
	}

	return result, nil
}

type StateBuilder struct {
	ClusterTemplate model.ClusterTemplate
	Stack           terra.Stack
	StateMgr        statemgr.Full
	Terra           *terra.Terra
}

func (sb StateBuilder) Build() (model.ClusterState, error) {
	if err := sb.StateMgr.RefreshState(); err != nil {
		return model.ClusterState{}, fmt.Errorf("sb.StateMgr.RefreshState: %w", err)
	}

	switch sb.ClusterTemplate.CloudProvider {
	case model.CloudProviderAWS:
		result, err := sb.buildAWS()
		if err != nil {
			return model.ClusterState{}, fmt.Errorf("sb.buildAWS: %w", err)
		}
		return result, nil
	case model.CloudProviderGCP:
		result, err := sb.buildGCP()
		if err != nil {
			return model.ClusterState{}, fmt.Errorf("sb.buildGCP: %w", err)
		}
		return result, nil
	default:
		return model.ClusterState{}, fmt.Errorf("unsupported cloud provider: %q",
			sb.ClusterTemplate.CloudProvider)
	}
}

func (sb StateBuilder) buildAWS() (model.ClusterState, error) {
	result := model.ClusterState{
		Template: sb.ClusterTemplate,
	}

	st := sb.StateMgr.State()

	for i, s := range sb.ClusterTemplate.Servers {
		res, err := sb.Terra.Evaluate(sb.Stack, st, addrs.RootModuleInstance.ResourceInstance(
			addrs.ManagedResourceMode, "aws_instance", s.ResourceID, addrs.NoKey))
		if err != nil {
			return result, fmt.Errorf("sb.Terra.Evaluate: %w", err)
		}

		pip := res.GetAttr("public_ip")
		if pip.IsNull() {
			return result, fmt.Errorf("public_ip is null for server %q", s.ResourceID)
		}

		id := res.GetAttr("id")
		if id.IsNull() {
			return result, fmt.Errorf("id is null for server %q", s.ResourceID)
		}

		ss := model.ServerState{
			Template: sb.ClusterTemplate.Servers[i],
			PublicIP: pip.AsString(),
			AWS: model.ServerStateAWS{
				ID: id.AsString(),
			},
		}
		result.Servers = append(result.Servers, ss)
	}

	return result, nil
}

func (sb StateBuilder) buildGCP() (model.ClusterState, error) {
	result := model.ClusterState{
		Template: sb.ClusterTemplate,
	}

	st := sb.StateMgr.State()

	for i, s := range sb.ClusterTemplate.Servers {
		res, err := sb.Terra.Evaluate(sb.Stack, st, addrs.RootModuleInstance.ResourceInstance(
			addrs.ManagedResourceMode, "google_compute_instance", s.ResourceID, addrs.NoKey))
		if err != nil {
			return result, fmt.Errorf("sb.Terra.Evaluate: %w", err)
		}

		pip, err := cty.GetAttrPath("network_interface").Index(cty.NumberIntVal(0)).
			GetAttr("access_config").Index(cty.NumberIntVal(0)).GetAttr("nat_ip").Apply(res)

		if err != nil {
			return result, fmt.Errorf("nat_ip: %w", err)
		}

		if pip.IsNull() {
			return result, fmt.Errorf("network_interface[0].access_config[0].nat_ip is null for server %q", s.ResourceID)
		}

		id := res.GetAttr("instance_id")
		if id.IsNull() {
			return result, fmt.Errorf("id is null for server %q", s.ResourceID)
		}

		selfLink := res.GetAttr("self_link")
		if selfLink.IsNull() {
			return result, fmt.Errorf("self_link is null for server %q", s.ResourceID)
		}

		ss := model.ServerState{
			Template: sb.ClusterTemplate.Servers[i],
			PublicIP: pip.AsString(),
			GCP: model.ServerStateGCP{
				SelfLink:   selfLink.AsString(),
				InstanceID: id.AsString(),
			},
		}
		result.Servers = append(result.Servers, ss)
	}

	return result, nil
}
