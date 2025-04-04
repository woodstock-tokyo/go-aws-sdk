package rds

import (
	goctx "context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

const (
	DefaultDBClusterParameterGroupName = "woodstock-aurora-cluster-prod"
	DefaultDBSubnetGroupName           = "main_subnet_group_prod_ap-northeast-1"
	DefaultEngine                      = "aurora-mysql"
)

type DBSnapshotsType string

const (
	DBSnapshotsTypeAutomated DBSnapshotsType = "automated"
	DBSnapshotsTypeManual    DBSnapshotsType = "manual"
)

// DescribeDBClusterSnapshotsOpts describe db snapshot options
type DescribeDBClusterSnapshotsOpts struct {
	DBClusterIdentifier string
	SnapshotType        DBSnapshotsType
	Timeout             time.Duration
}

// DescribeDBClusterSnapshotsResponse describe db snapshot response
type DescribeDBClusterSnapshotsResponse struct {
	DBClusterIdentifier string
	Snapshots           []*rds.DBClusterSnapshot
	Error               error
}

// RestoreDBClusterFromSnapshotOpts restore db instance from db snapshot options
type RestoreDBClusterFromSnapshotOpts struct {
	DBClusterIdentifier string
	SnapshotIdentifier  string
	Timeout             time.Duration
}

// RestoreDBClusterFromSnapshotResponse restore db instance from db snapshot response
type RestoreDBClusterFromSnapshotResponse struct {
	Error error
}

// Context context includes endpoint, region and bucket info
type context struct {
	region string
}

// Service service includes context and credentials
type Service struct {
	context      *context
	accessKey    string
	accessSecret string
}

// NewService service initializer
func NewService(key, secret string) *Service {
	return &Service{
		context:      new(context),
		accessKey:    key,
		accessSecret: secret,
	}
}

// SetRegion set region
func (s *Service) SetRegion(region string) {
	s.context.check()
	s.context.region = region
}

// GetRegion get region
func (s *Service) GetRegion() string {
	return s.context.region
}

var once sync.Once
var instance *rds.RDS

// client init client
func (s *Service) client() *rds.RDS {
	once.Do(func() {
		sess, _ := session.NewSession(&aws.Config{
			Region:      aws.String(s.GetRegion()),
			Credentials: credentials.NewStaticCredentials(s.accessKey, s.accessSecret, ""),
		})

		instance = rds.New(sess)
	})

	return instance
}

// DescribeDBSnapshots describe db snapshots
func (s *Service) DescribeDBSnapshots(opts *DescribeDBClusterSnapshotsOpts) (resp *DescribeDBClusterSnapshotsResponse) {
	s.context.check()
	resp = &DescribeDBClusterSnapshotsResponse{
		Snapshots: []*rds.DBClusterSnapshot{},
	}

	client := s.client()
	t := 180 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}

	snapshotType := string(DBSnapshotsTypeAutomated)
	if opts.SnapshotType == DBSnapshotsTypeManual {
		snapshotType = string(DBSnapshotsTypeManual)
	}

	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	snapshotResp, err := client.DescribeDBClusterSnapshotsWithContext(ctx, &rds.DescribeDBClusterSnapshotsInput{
		DBClusterIdentifier: aws.String(opts.DBClusterIdentifier),
		SnapshotType:        aws.String(snapshotType),
	})

	if err != nil {
		resp.Error = err
	} else {
		resp.Snapshots = snapshotResp.DBClusterSnapshots
	}

	return
}

// DescribeLatestDBSnapshot describe latest db snapshot
func (s *Service) DescribeLatestDBSnapshot(opts *DescribeDBClusterSnapshotsOpts) (snapshot *rds.DBClusterSnapshot, err error) {
	snapshotsResp := s.DescribeDBSnapshots(opts)
	if snapshotsResp.Error != nil {
		return nil, snapshotsResp.Error
	}

	if len(snapshotsResp.Snapshots) == 0 {
		return nil, nil
	}

	sort.Slice(snapshotsResp.Snapshots, func(i, j int) bool {
		return snapshotsResp.Snapshots[i].SnapshotCreateTime.After(*snapshotsResp.Snapshots[j].SnapshotCreateTime)
	})

	return snapshotsResp.Snapshots[0], nil
}

// RestoreDBClusterFromSnapshot restore db instance from db snapshot
func (s *Service) RestoreDBClusterFromSnapshot(opts *RestoreDBClusterFromSnapshotOpts) (resp *RestoreDBClusterFromSnapshotResponse) {
	s.context.check()
	if opts.DBClusterIdentifier == "woodstock-prod" {
		return &RestoreDBClusterFromSnapshotResponse{
			Error: fmt.Errorf("woodstock-prod db cluster cannot be restored"),
		}
	}

	client := s.client()
	t := 30 * 60 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}

	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	_, err := client.RestoreDBClusterFromSnapshotWithContext(ctx, &rds.RestoreDBClusterFromSnapshotInput{
		DBClusterIdentifier:         aws.String(opts.DBClusterIdentifier),
		SnapshotIdentifier:          aws.String(opts.SnapshotIdentifier),
		DBClusterParameterGroupName: aws.String(DefaultDBClusterParameterGroupName),
		DBSubnetGroupName:           aws.String(DefaultDBSubnetGroupName),
		Engine:                      aws.String(DefaultEngine),
	})

	if err != nil {
		resp.Error = err
	}

	return
}

func (c *context) check() {
	if c == nil {
		panic("invalid context")
	}
}
