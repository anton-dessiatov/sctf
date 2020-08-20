package dal

import (
	"fmt"
	"time"

	"github.com/anton-dessiatov/sctf/tf/model"
	"github.com/jinzhu/gorm"
)

type Cluster struct {
	ID        int
	Template  model.ClusterTemplate
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ClusterByID(db *gorm.DB, clusterID int) (Cluster, error) {
	var cluster Cluster
	if err := db.Where("id = ?", clusterID).First(&cluster).Error; err != nil {
		return Cluster{}, fmt.Errorf("app.Instance.DB.Where: %w", err)
	}

	return cluster, nil
}

type Stack struct {
	ID        int
	ClusterID int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type State struct {
	ID        int
	StackID   int
	Body      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}
