package cluster

import (
	"fmt"
	"log"

	"github.com/anton-dessiatov/sctf/direct/app"
	"github.com/anton-dessiatov/sctf/direct/dal"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jinzhu/gorm"
)

func CmdCreateEIP(app *app.App, clusterID int) error {
	var c dal.Cluster
	if err := app.DB.Where("id = ?", clusterID).First(&c).Error; err != nil {
		return fmt.Errorf("app.DB.First: %w", err)
	}

	var instances []dal.Instance
	if err := app.DB.Where("cluster_id = ? AND status = ?", clusterID, "ACTIVE").Find(&instances).Error; err != nil {
		return fmt.Errorf("app.DB.Find: %w", err)
	}

	createEIP := NewCreateEIP(app.DB, ec2.New(app.AWS), &c)
	for i := range instances {
		_, err := createEIP.Do(&instances[i])
		if err != nil {
			return fmt.Errorf("createEIP.Do: %w", err)
		}
	}

	return nil
}

type CreateEIP struct {
	db  *gorm.DB
	ec2 *ec2.EC2

	cluster *dal.Cluster
}

func NewCreateEIP(db *gorm.DB, ec2 *ec2.EC2, cluster *dal.Cluster) *CreateEIP {
	return &CreateEIP{
		db:  db,
		ec2: ec2,

		cluster: cluster,
	}
}

func (c *CreateEIP) Do(instance *dal.Instance) (*dal.ElasticIP, error) {
	allocateAddressOutput, err := c.ec2.AllocateAddress(&ec2.AllocateAddressInput{})
	if err != nil {
		return nil, fmt.Errorf("c.ec2.AllocatedAddress: %w", err)
	}

	log.Printf("Allocated address %q: %q", *allocateAddressOutput.AllocationId,
		*allocateAddressOutput.PublicIp)

	eip := dal.ElasticIP{
		ClusterID:  c.cluster.ID,
		ExternalID: *allocateAddressOutput.AllocationId,
		Status:     "ACTIVE",
	}
	if err := c.db.Create(&eip).Error; err != nil {
		return nil, fmt.Errorf("c.db.Create: %w", err)
	}

	associateAddressOutput, err := c.ec2.AssociateAddress(&ec2.AssociateAddressInput{
		AllocationId: aws.String(eip.ExternalID),
		InstanceId:   aws.String(instance.ExternalID),
	})
	if err != nil {
		return nil, fmt.Errorf("c.ec2.AssociateAddress: %w", err)
	}

	log.Printf("Associated address %q(%q) with instance %q: %q",
		*allocateAddressOutput.AllocationId, *allocateAddressOutput.PublicIp,
		instance.ExternalID, *associateAddressOutput.AssociationId)

	eip.InstanceID = &instance.ID
	eip.AssociationID = associateAddressOutput.AssociationId
	if err := c.db.Save(&eip).Error; err != nil {
		return nil, fmt.Errorf("c.db.Update: %w", err)
	}

	instance.PublicIP = allocateAddressOutput.PublicIp
	if err := c.db.Save(instance).Error; err != nil {
		return nil, fmt.Errorf("c.db.Update: %w", err)
	}

	return &eip, nil
}
