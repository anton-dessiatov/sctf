package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/direct/app"
	"github.com/anton-dessiatov/sctf/direct/dal"
	"github.com/anton-dessiatov/sctf/direct/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func Create(app *app.App, template model.ClusterTemplate) error {
	op, err := newCreateOperation(app)
	if err != nil {
		return fmt.Errorf("newCreateOperation: %w", err)
	}

	if err := op.createVpc(template.PrivateCIDR); err != nil {
		return fmt.Errorf("op.createVpc: %w", err)
	}

	if err := op.createInternetGateway(); err != nil {
		return fmt.Errorf("op.createInternetGateway: %w", err)
	}

	if err := op.createSecurityGroups(); err != nil {
		return fmt.Errorf("op.createSecurityGroups: %w", err)
	}

	for i := range template.Servers {
		subnet, err := op.createSubnet(template.Servers[i].AvailabilityZone,
			template.Servers[i].SubnetCIDR)
		if err != nil {
			return fmt.Errorf("op.createSubnet: %w", err)
		}

		err = op.createRtb(subnet)
		if err != nil {
			return fmt.Errorf("op.createRtb: %w", err)
		}

		instance, err := op.createInstance(subnet)
		if err != nil {
			return fmt.Errorf("op.createInstance: %w", err)
		}

		err = op.createEIP(instance)
		if err != nil {
			return fmt.Errorf("op.createEIP: %w", err)
		}
	}

	op.cluster.Status = "ACTIVE"
	if err := op.app.DB.Save(op.cluster).Error; err != nil {
		return fmt.Errorf("op.app.DB.Save: %w", err)
	}

	log.Printf("Created cluster ID %d: %+v\n", op.cluster.ID, op.cluster)

	return nil
}

type createOperation struct {
	app     *app.App
	cluster *dal.Cluster
	vpc     *dal.VPC
	igw     *dal.IGW

	securityGroups map[string]*dal.SecurityGroup
	subnets        []*dal.Subnet
	rtbs           []*dal.RTB
	instances      []*dal.Instance
	eips           []*dal.ElasticIP

	ec2 *ec2.EC2
}

func newCreateOperation(app *app.App) (*createOperation, error) {
	result := createOperation{
		app: app,
		cluster: &dal.Cluster{
			Status: "CREATING",
		},

		securityGroups: make(map[string]*dal.SecurityGroup),

		ec2: ec2.New(app.AWS),
	}

	if err := app.DB.Create(result.cluster).Error; err != nil {
		return nil, fmt.Errorf("app.DB.Create: %w", err)
	}

	return &result, nil
}

func (op *createOperation) createVpc(cidrBlock string) error {
	createVpcOutput, err := op.ec2.CreateVpc(&ec2.CreateVpcInput{
		CidrBlock: aws.String(cidrBlock),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.CreateVpc: %w", err)
	}
	log.Printf("Created VPC %q\n", *createVpcOutput.Vpc.VpcId)

	op.vpc = &dal.VPC{
		ClusterID:  op.cluster.ID,
		ExternalID: *createVpcOutput.Vpc.VpcId,
		Status:     "ACTIVE",
	}

	if err := op.app.DB.Create(op.vpc).Error; err != nil {
		return fmt.Errorf("app.DB.Create: %w", err)
	}

	return nil
}

func (op *createOperation) createInternetGateway() error {
	createIgwOutput, err := op.ec2.CreateInternetGateway(&ec2.CreateInternetGatewayInput{})
	if err != nil {
		return fmt.Errorf("op.ec2.CreateInternetGateway: %w", err)
	}
	log.Printf("Created IGW %q\n", *createIgwOutput.InternetGateway.InternetGatewayId)

	op.igw = &dal.IGW{
		ClusterID:  op.cluster.ID,
		ExternalID: *createIgwOutput.InternetGateway.InternetGatewayId,
		Status:     "UNATTACHED",
	}

	if err := op.app.DB.Create(op.igw).Error; err != nil {
		return fmt.Errorf("app.DB.Create: %w", err)
	}

	_, err = op.ec2.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: createIgwOutput.InternetGateway.InternetGatewayId,
		VpcId:             aws.String(op.vpc.ExternalID),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.AttachInternetGateway: %w", err)
	}

	op.igw.Status = "ATTACHED"
	if err := op.app.DB.Save(op.igw).Error; err != nil {
		return fmt.Errorf("app.DB.Save: %w", err)
	}

	return nil
}

func (op *createOperation) createSubnet(az, cidrBlock string) (*dal.Subnet, error) {
	createSubnetOutput, err := op.ec2.CreateSubnet(&ec2.CreateSubnetInput{
		AvailabilityZone: aws.String(az),
		CidrBlock:        aws.String(cidrBlock),
		VpcId:            aws.String(op.vpc.ExternalID),
	})
	if err != nil {
		return nil, fmt.Errorf("op.ec2.CreateSubnet: %w", err)
	}
	log.Printf("Created subnet %q (%q, %q)\n", *createSubnetOutput.Subnet.SubnetId,
		az, cidrBlock)

	subnet := dal.Subnet{
		ClusterID:  op.cluster.ID,
		ExternalID: *createSubnetOutput.Subnet.SubnetId,
		Status:     "ACTIVE",
	}
	if err := op.app.DB.Create(&subnet).Error; err != nil {
		return nil, fmt.Errorf("op.app.DB.Create: %w", err)
	}

	op.subnets = append(op.subnets, &subnet)

	return &subnet, nil
}

func (op *createOperation) createRtb(subnet *dal.Subnet) error {
	createRtbOutput, err := op.ec2.CreateRouteTable(&ec2.CreateRouteTableInput{
		VpcId: aws.String(op.vpc.ExternalID),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.CreateRouteTable: %w", err)
	}
	log.Printf("Created routing table %q\n", *createRtbOutput.RouteTable.RouteTableId)

	rtb := dal.RTB{
		ClusterID:  op.cluster.ID,
		ExternalID: *createRtbOutput.RouteTable.RouteTableId,
		Status:     "UNATTACHED",
	}
	if err := op.app.DB.Create(&rtb).Error; err != nil {
		return fmt.Errorf("op.app.DB.Create: %w", err)
	}
	op.rtbs = append(op.rtbs, &rtb)

	_, err = op.ec2.CreateRoute(&ec2.CreateRouteInput{
		RouteTableId:         aws.String(rtb.ExternalID),
		GatewayId:            aws.String(op.igw.ExternalID),
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.CreateRoute: %w", err)
	}
	log.Printf("Added route 0.0.0.0/0 via %q to %q\n", op.igw.ExternalID, rtb.ExternalID)

	associateRouteTableOutput, err := op.ec2.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(rtb.ExternalID),
		SubnetId:     aws.String(subnet.ExternalID),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.AssociateRouteTable: %w", err)
	}
	log.Printf("Associated route table %q with subnet %q: %q\n", rtb.ExternalID, subnet.ExternalID,
		*associateRouteTableOutput.AssociationId)

	rtb.SubnetID = &subnet.ID
	rtb.AssociationID = associateRouteTableOutput.AssociationId
	rtb.Status = "ATTACHED"

	if err := op.app.DB.Save(&rtb).Error; err != nil {
		return fmt.Errorf("op.app.DB.Save: %w", err)
	}

	return nil
}

func (op *createOperation) createSecurityGroups() error {
	if err := op.createSSHSecurityGroup(); err != nil {
		return fmt.Errorf("op.createSSHSecruityGroup: %w", err)
	}

	return nil
}

func (op *createOperation) createSSHSecurityGroup() error {
	createSgOutput, err := op.ec2.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String("allow_ssh"),
		VpcId:       aws.String(op.vpc.ExternalID),
		Description: aws.String("SSH access - one and for all"),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.CreateSecurityGroup: %w", err)
	}
	log.Printf("Created security group %q\n", *createSgOutput.GroupId)

	sg := dal.SecurityGroup{
		ClusterID:  op.cluster.ID,
		Kind:       "ALLOW_SSH",
		ExternalID: *createSgOutput.GroupId,
		Status:     "PENDING",
	}
	if err := op.app.DB.Create(&sg).Error; err != nil {
		return fmt.Errorf("op.db.Create: %w", err)
	}
	op.securityGroups[sg.Kind] = &sg

	_, err = op.ec2.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		CidrIp:     aws.String("0.0.0.0/0"),
		FromPort:   aws.Int64(22),
		ToPort:     aws.Int64(22),
		GroupId:    aws.String(sg.ExternalID),
		IpProtocol: aws.String("tcp"),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.AuthorizeSecurityGroupIngress: %w", err)
	}
	sg.Status = "ACTIVE"
	if err := op.app.DB.Save(&sg).Error; err != nil {
		return fmt.Errorf("op.app.DB.Save: %w", err)
	}

	return nil
}

func (op *createOperation) createInstance(subnet *dal.Subnet) (*dal.Instance, error) {
	runInstancesOutput, err := op.ec2.RunInstances(&ec2.RunInstancesInput{
		InstanceType: aws.String("t2.micro"),
		ImageId:      aws.String("ami-0817d428a6fb68645"),
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:              aws.Int64(0),
				AssociatePublicIpAddress: aws.Bool(true),
				Groups: []*string{
					aws.String(op.securityGroups["ALLOW_SSH"].ExternalID),
				},
				SubnetId: aws.String(subnet.ExternalID),
			},
		},
		KeyName:  aws.String("anton_home"),
		MinCount: aws.Int64(1),
		MaxCount: aws.Int64(1),
	})
	if err != nil {
		return nil, fmt.Errorf("op2.ec2.RunInstances: %w", err)
	}
	log.Printf("Created instance %q\n", *runInstancesOutput.Instances[0].InstanceId)

	instance := dal.Instance{
		ClusterID:  op.cluster.ID,
		ExternalID: *runInstancesOutput.Instances[0].InstanceId,
		// PublicIP:   *runInstancesOutput.Instances[0].PublicIpAddress,
		Status: "PENDING",
	}
	if err := op.app.DB.Create(&instance).Error; err != nil {
		return nil, fmt.Errorf("op.app.DB.Create: %w", err)
	}
	op.instances = append(op.instances, &instance)

	log.Printf("Waiting until instance is RUNNING: %q\n", instance.ExternalID)

	describeInstancesInput := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-id"),
				Values: []*string{
					aws.String(instance.ExternalID),
				},
			},
		},
	}

	err = op.ec2.WaitUntilInstanceRunning(describeInstancesInput)
	if err != nil {
		return nil, fmt.Errorf("op.ec2.WaitUntilInstanceRunning: %w", err)
	}

	log.Printf("Waited successfully, querying instance %q public IP\n", instance.ExternalID)

	describeInstancesOut, err := op.ec2.DescribeInstances(describeInstancesInput)
	if err != nil {
		return nil, fmt.Errorf("op.ec2.DescribeInstances: %w", err)
	}

	publicIP := *describeInstancesOut.Reservations[0].Instances[0].PublicIpAddress
	log.Printf("Instance %q has public IP %q\n", instance.ExternalID, publicIP)

	instance.PublicIP = &publicIP
	instance.Status = "ACTIVE"

	if err := op.app.DB.Save(&instance).Error; err != nil {
		return nil, fmt.Errorf("op.app.DB.Save: %w", err)
	}

	return &instance, nil
}

func (op *createOperation) createEIP(instance *dal.Instance) error {
	eip, err := NewCreateEIP(op.app.DB, op.ec2, op.cluster).Do(instance)
	if err != nil {
		return fmt.Errorf("NewCreateEIP.Do: %w", err)
	}

	op.eips = append(op.eips, eip)

	return nil
}
