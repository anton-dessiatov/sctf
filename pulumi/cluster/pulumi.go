package cluster

import (
	"github.com/anton-dessiatov/sctf/pulumi/app"
)

func Up(app *app.App, clusterID int) error {
	// ws := ClusterWorkspace{
	// 	App:       app,
	// 	ClusterID: clusterID,
	// }
	// stack, err := auto.NewStackInlineSource(ctx, "myOrg/myProj/myStack", func(pCtx *pulumi.Context) error {
	// 	_, err := ec2.NewVpc(pCtx, "main", &ec2.VpcArgs{
	// 		CidrBlock: pulumi.String("10.0.0.0/16"),
	// 	})
	// 	if err != nil {
	// 		return fmt.Errorf("ec2.NewVpc: %w", err)
	// 	}
	// 	return nil
	// })
	return nil
}
