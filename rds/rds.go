package rds

import (
	goctx "context"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

type DBSnapshotsType string

const (
	DBSnapshotsTypeAutomated DBSnapshotsType = "automated"
	DBSnapshotsTypeManual    DBSnapshotsType = "manual"
)

// DescribeDBSnapshotsOpts describe db snapshot options
type DescribeDBSnapshotsOpts struct {
	SnapshotType DBSnapshotsType
	Timeout      time.Duration
}

// DescribeDBSnapshotsResponse describe db snapshot response
type DescribeDBSnapshotsResponse struct {
	Snapshots []*rds.DBSnapshot
	Error     error
}

// RestoreDBInstanceFromDBSnapshotOpts restore db instance from db snapshot options
type RestoreDBInstanceFromDBSnapshotOpts struct {
	SnapshotIdentifier string
	Timeout            time.Duration
}

// RestoreDBInstanceFromDBSnapshotResponse restore db instance from db snapshot response
type RestoreDBInstanceFromDBSnapshotResponse struct {
	Error error
}

// Context context includes endpoint, region and bucket info
type context struct {
	identifier string
	region     string
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

// SetIdentifier set identifier
func (s *Service) SetIdentifier(identifier string) {
	s.context.check()
	s.context.identifier = identifier
}

// GetIdentifier get identifier
func (s *Service) GetIdentifier() string {
	return s.context.identifier
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
func (s *Service) DescribeDBSnapshots(opts *DescribeDBSnapshotsOpts) (resp *DescribeDBSnapshotsResponse) {
	s.context.check()
	resp = &DescribeDBSnapshotsResponse{
		Snapshots: []*rds.DBSnapshot{},
	}

	client := s.client()
	t := 180 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}

	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	snapshotResp, err := client.DescribeDBSnapshotsWithContext(ctx, &rds.DescribeDBSnapshotsInput{
		DBInstanceIdentifier: aws.String(s.context.identifier),
		SnapshotType:         aws.String("automated"),
	})

	if err != nil {
		resp.Error = err
	} else {
		resp.Snapshots = snapshotResp.DBSnapshots
	}

	return
}

// DescribeLatestDBSnapshot describe latest db snapshot
func (s *Service) DescribeLatestDBSnapshot(opts *DescribeDBSnapshotsOpts) (snapshot *rds.DBSnapshot, err error) {
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

// RestoreDBInstanceFromSnapshot restore db instance from db snapshot
func (s *Service) RestoreDBInstanceFromSnapshot(opts *RestoreDBInstanceFromDBSnapshotOpts) (resp *RestoreDBInstanceFromDBSnapshotResponse) {
	s.context.check()
	client := s.client()
	t := 180 * time.Second
	if opts.Timeout > 0 {
		t = opts.Timeout
	}

	ctx, cancel := goctx.WithTimeout(goctx.Background(), t)
	defer cancel()

	_, err := client.RestoreDBInstanceFromDBSnapshotWithContext(ctx, &rds.RestoreDBInstanceFromDBSnapshotInput{
		DBInstanceIdentifier: aws.String(s.context.identifier), // not the snapshot identifier, but the new instance identifier
		DBSnapshotIdentifier: aws.String(opts.SnapshotIdentifier),
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
