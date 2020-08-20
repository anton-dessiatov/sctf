package terra

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/configs"
	"github.com/spf13/afero"
)

type StackIdentity struct {
	ClusterID int
	Name      string
}

type Stack struct {
	AWS    ConfigAWS
	GCP    ConfigGCP
	Config StackModule
}

type ConfigAWS struct {
	Region string
}

type ConfigGCP struct {
	Region string
}

type StackModule interface {
	ToModule() (*configs.Module, error, hcl.Diagnostics)
}

type StackDirect struct {
	Resources map[string]*configs.Resource
}

func (s *StackDirect) ToModule() (*configs.Module, error, hcl.Diagnostics) {
	return &configs.Module{
		ManagedResources: s.Resources,
	}, nil, nil
}

type StackText string

func (s StackText) ToModule() (*configs.Module, error, hcl.Diagnostics) {
	fs, err := s.fs()
	if err != nil {
		return nil, fmt.Errorf("s.fs: %w", err), nil
	}

	p := configs.NewParser(fs)
	mod, diags := p.LoadConfigDir("")
	if diags.HasErrors() {
		return nil, nil, diags
	}

	return mod, nil, diags
}

func (s StackText) fs() (afero.Afero, error) {
	result := afero.Afero{Fs: afero.NewMemMapFs()}
	err := result.WriteFile("main.tf", []byte(s), 0644)
	if err != nil {
		return afero.Afero{}, fmt.Errorf("result.WriteFile: %w", err)
	}

	return result, nil
}
