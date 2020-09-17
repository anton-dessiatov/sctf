package cluster

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/anton-dessiatov/sctf/tf/dal"
	"github.com/anton-dessiatov/sctf/tf/model"
	"github.com/anton-dessiatov/sctf/tf/terra"
	"github.com/jinzhu/gorm"
)

func StackByClusterID(db *gorm.DB, clusterID int) (terra.Stack, error) {
	c, err := dal.ClusterByID(db, clusterID)
	if err != nil {
		return terra.Stack{}, fmt.Errorf("dal.ClusterByID: %w", err)
	}

	stack, err := Stack(model.ClusterIdentity(clusterID), c.Template)
	if err != nil {
		return terra.Stack{}, fmt.Errorf("cluster.Stack: %w", err)
	}

	return stack, nil
}

func StackIdentity(clusterID int) terra.StackIdentity {
	return terra.StackIdentity{
		ClusterID: clusterID,
		Name:      "primary",
	}
}

func Stack(id model.ClusterIdentity, cluster model.ClusterTemplate) (terra.Stack, error) {
	switch cluster.CloudProvider {
	case model.CloudProviderAWS:
		return stackAWS(id, cluster)
	case model.CloudProviderGCP:
		return stackGCP(id, cluster)
	default:
		return terra.Stack{}, fmt.Errorf("unsupported cloud provider: %q", cluster.CloudProvider)
	}
}

// Stack terraform definitions must be kept in sync with state builder implementation. Make sure
// that they are aligned (maybe we could implement automated ways of doing so? At least we could
// make it less tedious with attributes & reflection)

func stackAWS(id model.ClusterIdentity, cluster model.ClusterTemplate) (terra.Stack, error) {
	config, err := configFromTemplate(id, cluster, `
resource aws_vpc main {
  cidr_block = "{{.PrivateCIDR}}"
}

data aws_ami ubuntu {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource aws_security_group allow_ssh {
  name        = "allow_ssh"
  description = "Allow SSH inbound traffic"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "SSH from world"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "allow_ssh"
  }
}

resource aws_internet_gateway gw {
  vpc_id = aws_vpc.main.id
}

{{range $server := .Servers}}
resource aws_subnet "{{.ResourceID}}_sub" {
  vpc_id = aws_vpc.main.id
  availability_zone = "{{.AWS.AvailabilityZone}}"
  cidr_block = "{{.SubnetCIDR}}"
}

resource aws_route_table "{{.ResourceID}}_rtb" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }
}

resource aws_route_table_association "{{.ResourceID}}_route_world" {
  subnet_id      = aws_subnet.{{.ResourceID}}_sub.id
  route_table_id = aws_route_table.{{.ResourceID}}_rtb.id
}

resource aws_instance "{{.ResourceID}}" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t2.micro"
  key_name		= "anton_home"

  disable_api_termination = true

  subnet_id = aws_subnet.{{.ResourceID}}_sub.id
  vpc_security_group_ids = [aws_security_group.allow_ssh.id]
  associate_public_ip_address = true

  tags = {
    Name = "{{.ResourceID}}"
  }
}
{{end}}
`)
	if err != nil {
		return terra.Stack{}, fmt.Errorf("configFromTemplate: %w", err)
	}
	return terra.Stack{
		AWS: terra.ConfigAWS{
			Region: cluster.AWS.Region,
		},
		Config: config,
	}, nil
}

func stackGCP(id model.ClusterIdentity, cluster model.ClusterTemplate) (terra.Stack, error) {
	config, err := configFromTemplate(id, cluster, `
locals {
  ssh_key = <<EOF
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC00U2OSoIRYpA1toFHRtTrifAZx3b0pT/WmOSBUd2/60PQB7X8c4RoaCej1L08QrCsHzHAxvWhZvWNip0TvzWXwNWwui0JvPS5Ac1UJOiVULMsRFt7VnwUH2PSEZpxU84/Atrgsf2lRL6SJjUtkBnKNgOQC01dC66gZ5EeGnSFPt0Bu7uFoCTzF7oIBb0SgRx7hmjEKMqAOgwERw5M9lCb5FhKMRltOCCLfN9Rva5iIjLa10yyv3zoha+APl3nfONOdgMeJf0toxfmHaPRm0S66vp65RiBi8pxlEAKJc55MONw6hj6JTkIkZgWI5I+rVWzt4pCqA5a4M9+9G78hTWN
EOF
}

data google_compute_image ubuntu {
  family  = "ubuntu-1804-lts"
  project = "ubuntu-os-cloud"
}

resource google_compute_network vpc_network {
  name = "sctf-{{.ID}}-vpc-network"
}

resource google_compute_firewall allow_ssh {
  name    = "sctf-{{.ID}}-allow-ssh"
  network = google_compute_network.vpc_network.name

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}

{{range $server := .Servers}}
resource google_compute_subnetwork "{{.ResourceID}}_sub" {
  name   = "sctf-{{$.ID}}-sub-{{.ResourceID}}"
  network = google_compute_network.vpc_network.id

  ip_cidr_range = "{{$server.SubnetCIDR}}"
}

resource google_compute_instance "{{.ResourceID}}" {
  name   = "sctf-{{$.ID}}-{{.ResourceID}}"
  machine_type = "e2-small"
  zone = "{{$server.GCP.AvailabilityZone}}"

  boot_disk {
    initialize_params {
      image = data.google_compute_image.ubuntu.self_link
    }
  }

  metadata = {
    ssh-keys = "ubuntu:${local.ssh_key}"
  }

  network_interface {
    network = google_compute_network.vpc_network.id
    subnetwork = google_compute_subnetwork.{{.ResourceID}}_sub.id
    access_config {}
  }
}
{{end}}
	`)

	if err != nil {
		return terra.Stack{}, fmt.Errorf("configFromTemplate: %w", err)
	}

	return terra.Stack{
		GCP: terra.ConfigGCP{
			Region: cluster.GCP.Region,
		},
		Config: config,
	}, nil
}

func configFromTemplate(id model.ClusterIdentity, cluster model.ClusterTemplate,
	tmplText string) (terra.StackModule, error) {
	data := struct {
		ID model.ClusterIdentity
		model.ClusterTemplate
	}{
		ID:              id,
		ClusterTemplate: cluster,
	}
	tmpl, err := template.New("cluster").Parse(tmplText)
	if err != nil {
		return nil, fmt.Errorf("template.New.Parse: %w", err)
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, data); err != nil {
		return nil, fmt.Errorf("tmpl.Execute: %w", err)
	}

	return terra.StackText(string(result.Bytes())), nil
}
