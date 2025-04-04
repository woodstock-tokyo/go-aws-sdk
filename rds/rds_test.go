package rds

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/woodstock-tokyo/woodstock-utils"
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
	assert.Equal(t, Ptov(snapshot.DBClusterIdentifier), "woodstock-prod")
}

func TestRestoreDBClusterFromSnapshot(t *testing.T) {
	snapshot, err := svc.DescribeLatestDBSnapshot(&DescribeDBClusterSnapshotsOpts{
		DBClusterIdentifier: "woodstock-prod",
	})
	assert.NoError(t, err)

	resp := svc.RestoreDBClusterFromSnapshot(&RestoreDBClusterFromSnapshotOpts{
		DBClusterIdentifier: "woodstock-pre-prod",
		SnapshotIdentifier:  Ptov(snapshot.DBClusterSnapshotIdentifier),
	})
	assert.NoError(t, resp.Error)
}
