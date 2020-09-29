package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/direct/app"
	"github.com/anton-dessiatov/sctf/direct/dal"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func Destroy(app *app.App, cluster dal.Cluster) error {
	log.Printf("Destroying cluster %v\n", cluster)

	op, err := newDestroyOperation(app, cluster)
	if err != nil {
		return fmt.Errorf("newDestroyOperation: %w", err)
	}

	for i := range op.eips {
		if err := op.destroyEIP(&op.eips[i]); err != nil {
			return fmt.Errorf("op.destroyEIP: %w", err)
		}
	}

	for i := range op.instances {
		if err := op.destroyInstance(&op.instances[i]); err != nil {
			return fmt.Errorf("op.destroyInstances: %w", err)
		}
	}

	for i := range op.rtbs {
		if err := op.destroyRtb(&op.rtbs[i]); err != nil {
			return fmt.Errorf("op.destroyRtb: %w", err)
		}
	}

	for i := range op.subnets {
		if err := op.destroySubnet(&op.subnets[i]); err != nil {
			return fmt.Errorf("op.destroySubnet: %w", err)
		}
	}

	if err := op.destroySecurityGroups(); err != nil {
		return fmt.Errorf("op.destroySecurityGroups: %w", err)
	}

	if err := op.destroyInternetGateway(); err != nil {
		return fmt.Errorf("op.destroyInternetGateway: %w", err)
	}

	if err := op.destroyVpc(); err != nil {
		return fmt.Errorf("op.destroyVpc: %w", err)
	}

	cluster.Status = "DELETED"
	if err := app.DB.Save(&cluster).Error; err != nil {
		return fmt.Errorf("app.DB.Save: %w", err)
	}

	return nil
}

type destroyOperation struct {
	app     *app.App
	cluster dal.Cluster
	vpc     dal.VPC
	igw     dal.IGW

	securityGroups map[string]dal.SecurityGroup
	subnets        []dal.Subnet
	rtbs           []dal.RTB
	instances      []dal.Instance
	eips           []dal.ElasticIP

	ec2 *ec2.EC2
}

func newDestroyOperation(app *app.App, cluster dal.Cluster) (*destroyOperation, error) {
	result := destroyOperation{
		app:     app,
		cluster: cluster,

		securityGroups: make(map[string]dal.SecurityGroup),

		ec2: ec2.New(app.AWS),
	}

	withClusterID := app.DB.Where("cluster_id = ?", cluster.ID)

	if err := withClusterID.Where("status = ?", "ACTIVE").First(&result.vpc).Error; err != nil {
		return nil, fmt.Errorf("withClusterID.Where.First(vpc): %w", err)
	}

	if err := withClusterID.Where("status = ? OR status = ?", "ATTACHED", "UNATTACHED").
		First(&result.igw).Error; err != nil {
		return nil, fmt.Errorf("withClusterID.Where.First(igw): %w", err)
	}

	var securityGroups []dal.SecurityGroup
	if err := withClusterID.Where("status = ?", "ACTIVE").Find(&securityGroups).Error; err != nil {
		return nil, fmt.Errorf("withClusterID.Where.Find(securityGroup): %w", err)
	}

	for i := range securityGroups {
		result.securityGroups[securityGroups[i].Kind] = securityGroups[i]
	}

	if err := withClusterID.Where("status = ?", "ACTIVE").Find(&result.subnets).Error; err != nil {
		return nil, fmt.Errorf("withClusterID.Where.Find(subnets): %w", err)
	}

	if err := withClusterID.Where("status = ? OR status = ?", "ATTACHED", "UNATTACHED").
		Find(&result.rtbs).Error; err != nil {
		return nil, fmt.Errorf("withClusterID.Where.Find(rtbs): %w", err)
	}

	if err := withClusterID.Where("status = ?", "ACTIVE").Find(&result.instances).Error; err != nil {
		return nil, fmt.Errorf("withClusterID.Where.Find(instances): %w")
	}

	if err := withClusterID.Where("status = ?", "ACTIVE").Find(&result.eips).Error; err != nil {
		return nil, fmt.Errorf("withClusterID.Where.Find(eips): %w", err)
	}

	cluster.Status = "DELETING"
	if err := app.DB.Save(&cluster).Error; err != nil {
		return nil, fmt.Errorf("app.DB.Update: %w", err)
	}

	return &result, nil
}

func (op *destroyOperation) destroyVpc() error {
	_, err := op.ec2.DeleteVpc(&ec2.DeleteVpcInput{
		VpcId: aws.String(op.vpc.ExternalID),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.DeleteVpc: %w", err)
	}
	log.Printf("Deleted VPC %q\n", op.vpc.ExternalID)

	op.vpc.Status = "DELETED"
	if err := op.app.DB.Save(&op.vpc).Error; err != nil {
		return fmt.Errorf("op.app.DB.Save: %w", err)
	}

	return nil
}

func (op *destroyOperation) destroyInternetGateway() error {
	if op.igw.Status == "ATTACHED" {
		_, err := op.ec2.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(op.igw.ExternalID),
			VpcId:             aws.String(op.vpc.ExternalID),
		})
		if err != nil {
			return fmt.Errorf("op.ec2.DetachInternetGateway: %w", err)
		}
		log.Printf("Detached internet gateway %q from %q", op.igw.ExternalID, op.vpc.ExternalID)

		op.igw.Status = "UNATTACHED"
		if err := op.app.DB.Save(&op.igw).Error; err != nil {
			return fmt.Errorf("op.app.DB.Save: %w", err)
		}
	}

	_, err := op.ec2.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(op.igw.ExternalID),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.DeleteInternetGateway: %w", err)
	}
	log.Printf("Deleted internet gateway %q\n", op.igw.ExternalID)

	op.igw.Status = "DELETED"
	if err := op.app.DB.Save(&op.igw).Error; err != nil {
		return fmt.Errorf("op.app.DB.Save: %w", err)
	}

	return nil
}

func (op *destroyOperation) destroySubnet(subnet *dal.Subnet) error {
	_, err := op.ec2.DeleteSubnet(&ec2.DeleteSubnetInput{
		SubnetId: aws.String(subnet.ExternalID),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.DeleteSubnet: %w", err)
	}
	log.Printf("Deleted subnet %q\n", subnet.ExternalID)

	subnet.Status = "DELETED"
	if err := op.app.DB.Save(subnet).Error; err != nil {
		return fmt.Errorf("op.app.DB.Save: %w", err)
	}

	return nil
}

func (op *destroyOperation) destroyRtb(rtb *dal.RTB) error {
	if rtb.Status == "ATTACHED" {
		_, err := op.ec2.DisassociateRouteTable(&ec2.DisassociateRouteTableInput{
			AssociationId: rtb.AssociationID,
		})
		if err != nil {
			return fmt.Errorf("op.ec2.DisassociateRouteTable: %w", err)
		}
		log.Printf("Disassociated route table %q (%q)\n", rtb.ExternalID, *rtb.AssociationID)

		rtb.SubnetID = nil
		rtb.AssociationID = nil
		rtb.Status = "UNATTACHED"
		if err := op.app.DB.Save(rtb).Error; err != nil {
			return fmt.Errorf("op.app.DB.Save: %w", err)
		}
	}

	_, err := op.ec2.DeleteRouteTable(&ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(rtb.ExternalID),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.DeleteRouteTable: %w", err)
	}
	log.Printf("Deleted route table %q\n", rtb.ExternalID)

	rtb.Status = "DELETED"
	if err := op.app.DB.Save(rtb).Error; err != nil {
		return fmt.Errorf("op.app.DB.Save: %w", err)
	}

	return nil
}

func (op *destroyOperation) destroySecurityGroups() error {
	for kind, sg := range op.securityGroups {
		_, err := op.ec2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(sg.ExternalID),
		})
		if err != nil {
			return fmt.Errorf("op.ec2.DeleteSecurityGroup(%s): %w", kind, err)
		}
		log.Printf("Deleted security group %q\n", sg.ExternalID)

		sg.Status = "DELETED"
		if err := op.app.DB.Save(sg).Error; err != nil {
			return fmt.Errorf("op.app.DB.Save: %w", err)
		}
	}

	return nil
}

func (op *destroyOperation) destroyInstance(instance *dal.Instance) error {
	_, err := op.ec2.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.ExternalID),
		},
	})
	if err != nil {
		return fmt.Errorf("op.ec2.TerminateInstances: %w", err)
	}
	log.Printf("Terminated instance %q\n", instance.ExternalID)

	instance.Status = "DELETING"
	if err := op.app.DB.Save(instance).Error; err != nil {
		return fmt.Errorf("op.app.DB.Save: %w", err)
	}

	log.Printf("Waiting until %q actually terminates\n", instance.ExternalID)

	err = op.ec2.WaitUntilInstanceTerminated(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-id"),
				Values: []*string{
					aws.String(instance.ExternalID),
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("op.ec2.WaitUntilInstanceTerminated: %w", err)
	}

	log.Printf("%q terminated\n", instance.ExternalID)

	instance.Status = "DELETED"
	if err := op.app.DB.Save(instance).Error; err != nil {
		return fmt.Errorf("op.app.DB.Save: %w", err)
	}

	return nil
}

func (op *destroyOperation) destroyEIP(eip *dal.ElasticIP) error {
	if eip.AssociationID != nil {
		_, err := op.ec2.DisassociateAddress(&ec2.DisassociateAddressInput{
			AssociationId: eip.AssociationID,
		})
		if err != nil {
			return fmt.Errorf("op.ec2.DisassociateAddress: %w", err)
		}

		log.Printf("Disassociated address %q (%q)", eip.ExternalID,
			*eip.AssociationID)

		eip.InstanceID = nil
		eip.AssociationID = nil
		if err := op.app.DB.Save(eip).Error; err != nil {
			return fmt.Errorf("op.app.DB.Update: %w", err)
		}
	}

	_, err := op.ec2.ReleaseAddress(&ec2.ReleaseAddressInput{
		AllocationId: aws.String(eip.ExternalID),
	})
	if err != nil {
		return fmt.Errorf("op.ec2.ReleaseAddress: %w", err)
	}
	log.Printf("Released address %q", eip.ExternalID)

	eip.Status = "DELETED"
	if err := op.app.DB.Save(eip).Error; err != nil {
		return fmt.Errorf("op.app.DB.Save: %w", err)
	}

	return nil
}
