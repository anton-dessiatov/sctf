package terra

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	tfplugin "github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/tfdiags"
)

func (t *Terra) providers() map[addrs.Provider]providers.Factory {
	return map[addrs.Provider]providers.Factory{
		addrs.NewDefaultProvider("aws"): func() (providers.Interface, error) {
			opts := &hclog.LoggerOptions{
				Name:  "plugin",
				Level: hclog.Trace,
			}
			if PluginLogging {
				opts.Output = os.Stderr
			} else {
				opts.Output = ioutil.Discard
			}
			logger := hclog.New(opts)

			config := &plugin.ClientConfig{
				HandshakeConfig:  tfplugin.Handshake,
				Logger:           logger,
				AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
				Managed:          true,
				Cmd:              exec.Command("terraform-provider-aws"),
				AutoMTLS:         false,
				VersionedPlugins: tfplugin.VersionedPlugins,
			}

			client := plugin.NewClient(config)
			rpcClient, err := client.Client()
			if err != nil {
				return nil, err
			}

			raw, err := rpcClient.Dispense(tfplugin.ProviderPluginName)
			if err != nil {
				return nil, err
			}

			// store the client so that the plugin can kill the child process
			p := raw.(*tfplugin.GRPCProvider)
			p.PluginClient = client

			return p, nil
		},
		addrs.NewDefaultProvider("google"): func() (providers.Interface, error) {
			opts := &hclog.LoggerOptions{
				Name:  "plugin",
				Level: hclog.Trace,
			}
			if PluginLogging {
				opts.Output = os.Stderr
			} else {
				opts.Output = ioutil.Discard
			}
			logger := hclog.New(opts)

			config := &plugin.ClientConfig{
				HandshakeConfig:  tfplugin.Handshake,
				Logger:           logger,
				AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
				Managed:          true,
				Cmd:              exec.Command("terraform-provider-google"),
				AutoMTLS:         false,
				VersionedPlugins: tfplugin.VersionedPlugins,
			}

			client := plugin.NewClient(config)
			rpcClient, err := client.Client()
			if err != nil {
				return nil, err
			}

			raw, err := rpcClient.Dispense(tfplugin.ProviderPluginName)
			if err != nil {
				return nil, err
			}

			// store the client so that the plugin can kill the child process
			p := raw.(*tfplugin.GRPCProvider)
			p.PluginClient = client

			return p, nil
		},
	}
}

func (t *Terra) providerConfigs(aws ConfigAWS, gcp ConfigGCP) (result map[string]*configs.Provider, diags tfdiags.Diagnostics) {
	result = make(map[string]*configs.Provider)

	if (aws != ConfigAWS{}) {
		awsConfig, awsDiags := t.awsConfig(aws)
		diags = diags.Append(awsDiags)
		if awsDiags.HasErrors() {
			return nil, diags
		}
		result["aws"] = &configs.Provider{
			Name:   "aws",
			Config: awsConfig,
		}
	}
	if (gcp != ConfigGCP{}) {
		gcpConfig, gcpDiags := t.gcpConfig(gcp)
		diags = diags.Append(gcpDiags)
		if gcpDiags.HasErrors() {
			return nil, diags
		}
		result["google"] = &configs.Provider{
			Name:   "google",
			Config: gcpConfig,
		}
	}

	return result, diags
}

func (t *Terra) awsConfig(c ConfigAWS) (result hcl.Body, diags tfdiags.Diagnostics) {
	hclText, err := template.New("").Parse(string(`
		region = "{{.Region}}"
		access_key = "{{.AWS.AccessKey}}"
		secret_key = "{{.AWS.SecretKey}}"
		{{ if .AssumeRoleARN }}
		assume_role{
			role_arn = "{{.AWS.AssumeRoleARN}}"
			external_id = "{{.AWS.AssumeRoleExternalID}}"
		}
		{{ end }}
	`))
	if err != nil {
		diags = diags.Append(fmt.Errorf("template.New.Parse: %v", err))
		return nil, diags
	}

	var b bytes.Buffer
	v := struct {
		Region string
		AWS
	}{
		Region: c.Region,
		AWS:    t.credentials.AWS,
	}
	if err := hclText.Execute(&b, v); err != nil {
		diags = diags.Append(fmt.Errorf("tmpl.Execute: %v", err))
		return nil, diags
	}

	f, parseDiags := hclsyntax.ParseConfig(b.Bytes(), "", hcl.Pos{})
	diags = diags.Append(parseDiags)
	if parseDiags.HasErrors() {
		return nil, diags
	}

	return f.Body, diags
}

func (t *Terra) gcpConfig(c ConfigGCP) (result hcl.Body, diags tfdiags.Diagnostics) {
	hclText, err := template.New("").Parse(string(`
		credentials = <<EOF
{{.GCP.JsonKey}}
EOF
		region = "{{.Region}}"
		project = "{{.GCP.Project}}"
	`))
	if err != nil {
		diags = diags.Append(fmt.Errorf("template.New.Parse: %v", err))
		return nil, diags
	}

	var b bytes.Buffer
	v := struct {
		Region string
		GCP
	}{
		Region: c.Region,
		GCP:    t.credentials.GCP,
	}
	if err := hclText.Execute(&b, v); err != nil {
		diags = diags.Append(fmt.Errorf("tmpl.Execute: %v", err))
		return nil, diags
	}

	f, parseDiags := hclsyntax.ParseConfig(b.Bytes(), "", hcl.Pos{})
	diags = diags.Append(parseDiags)
	if parseDiags.HasErrors() {
		return nil, diags
	}

	return f.Body, diags
}
