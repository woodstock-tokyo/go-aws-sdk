package rds

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
)

var svc *Service

func init() {
	svc = NewService(os.Getenv("WS_RDS_AWS_ACCESS_KEY_ID"), os.Getenv("WS_RDS_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
}

// DescribeLatestDBSnapshot test describe lastest db snapshot
func TestDescribeLatestDBSnapshot(t *testing.T) {
	snapshot, err := svc.DescribeLatestDBSnapshot(&DescribeDBClusterSnapshotsOpts{
		DBClusterIdentifier: "woodstock-prod",
	})

	assert.NoError(t, err)
	assert.Equal(t, aws.StringValue(snapshot.DBClusterIdentifier), "woodstock-prod")
}

func TestRestoreDBClusterFromSnapshot(t *testing.T) {
	snapshot, err := svc.DescribeLatestDBSnapshot(&DescribeDBClusterSnapshotsOpts{
		DBClusterIdentifier: "woodstock-prod",
	})
	assert.NoError(t, err)

	resp := svc.RestoreDBClusterFromSnapshot(&RestoreDBClusterFromSnapshotOpts{
		DBClusterIdentifier: "woodstock-pre-prod",
		SnapshotIdentifier:  aws.StringValue(snapshot.DBClusterSnapshotIdentifier),
	})
	assert.NoError(t, resp.Error)
}
