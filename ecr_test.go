package ecrauditor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

type fakeECRClient struct {
	repositories       []types.Repository
	images             map[string][]types.ImageDetail
	lifecyclePolicies  map[string]*ecr.GetLifecyclePolicyOutput
	describeReposErr   error
	describeImagesErr  error
	getLifecycleErr    error
}

func (f *fakeECRClient) DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	if f.describeReposErr != nil {
		return nil, f.describeReposErr
	}
	return &ecr.DescribeRepositoriesOutput{
		Repositories: f.repositories,
	}, nil
}

func (f *fakeECRClient) DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error) {
	if f.describeImagesErr != nil {
		return nil, f.describeImagesErr
	}
	repoName := aws.ToString(params.RepositoryName)
	images, ok := f.images[repoName]
	if !ok {
		images = []types.ImageDetail{}
	}
	return &ecr.DescribeImagesOutput{
		ImageDetails: images,
	}, nil
}

func (f *fakeECRClient) GetLifecyclePolicy(ctx context.Context, params *ecr.GetLifecyclePolicyInput, optFns ...func(*ecr.Options)) (*ecr.GetLifecyclePolicyOutput, error) {
	if f.getLifecycleErr != nil {
		return nil, f.getLifecycleErr
	}
	repoName := aws.ToString(params.RepositoryName)
	policy, ok := f.lifecyclePolicies[repoName]
	if !ok {
		return nil, errors.New("LifecyclePolicyNotFoundException")
	}
	return policy, nil
}

func TestAuditor_Collect(t *testing.T) {
	now := time.Now()
	client := &fakeECRClient{
		repositories: []types.Repository{
			{
				RepositoryName: aws.String("repo1"),
				RepositoryUri:  aws.String("123456789.dkr.ecr.us-east-1.amazonaws.com/repo1"),
				CreatedAt:      aws.Time(now),
			},
			{
				RepositoryName: aws.String("repo2"),
				RepositoryUri:  aws.String("123456789.dkr.ecr.us-east-1.amazonaws.com/repo2"),
				CreatedAt:      aws.Time(now),
			},
		},
		images: map[string][]types.ImageDetail{
			"repo1": {
				{
					ImageDigest:      aws.String("sha256:abc123"),
					ImageTags:        []string{"latest", "v1.0"},
					ImagePushedAt:    aws.Time(now),
					ImageSizeInBytes: aws.Int64(1024000),
				},
				{
					ImageDigest:      aws.String("sha256:def456"),
					ImageTags:        []string{"v0.9"},
					ImagePushedAt:    aws.Time(now),
					ImageSizeInBytes: aws.Int64(2048000),
				},
			},
			"repo2": {
				{
					ImageDigest:      aws.String("sha256:ghi789"),
					ImagePushedAt:    aws.Time(now),
					ImageSizeInBytes: aws.Int64(512000),
				},
			},
		},
		lifecyclePolicies: map[string]*ecr.GetLifecyclePolicyOutput{
			"repo1": {
				RegistryId:          aws.String("123456789"),
				RepositoryName:      aws.String("repo1"),
				LifecyclePolicyText: aws.String(`{"rules":[]}`),
			},
		},
	}

	auditor := NewAuditorWithClient(client, "us-east-1", "test-profile")
	result, err := auditor.Collect(context.Background())

	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	if result.Region != "us-east-1" {
		t.Errorf("Region = %s, expected us-east-1", result.Region)
	}
	if result.Profile != "test-profile" {
		t.Errorf("Profile = %s, expected test-profile", result.Profile)
	}
	if result.RepositoryCount != 2 {
		t.Errorf("RepositoryCount = %d, expected 2", result.RepositoryCount)
	}
	if len(result.Repositories) != 2 {
		t.Fatalf("Repositories length = %d, expected 2", len(result.Repositories))
	}

	repo1 := result.Repositories[0]
	if repo1.Name != "repo1" {
		t.Errorf("repo1.Name = %s, expected repo1", repo1.Name)
	}
	if repo1.ImageCount != 2 {
		t.Errorf("repo1.ImageCount = %d, expected 2", repo1.ImageCount)
	}
	if repo1.LifecyclePolicy == nil {
		t.Error("repo1 should have lifecycle policy")
	}

	repo2 := result.Repositories[1]
	if repo2.Name != "repo2" {
		t.Errorf("repo2.Name = %s, expected repo2", repo2.Name)
	}
	if repo2.ImageCount != 1 {
		t.Errorf("repo2.ImageCount = %d, expected 1", repo2.ImageCount)
	}
	if repo2.LifecyclePolicy != nil {
		t.Error("repo2 should not have lifecycle policy")
	}
}

func TestAuditor_Collect_NoRepositories(t *testing.T) {
	client := &fakeECRClient{
		repositories: []types.Repository{},
		images:       map[string][]types.ImageDetail{},
	}

	auditor := NewAuditorWithClient(client, "us-east-1", "test")
	result, err := auditor.Collect(context.Background())

	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if result.RepositoryCount != 0 {
		t.Errorf("RepositoryCount = %d, expected 0", result.RepositoryCount)
	}
}

func TestAuditor_Collect_DescribeReposError(t *testing.T) {
	expectedErr := errors.New("API error")
	client := &fakeECRClient{
		describeReposErr: expectedErr,
	}

	auditor := NewAuditorWithClient(client, "us-east-1", "test")
	_, err := auditor.Collect(context.Background())

	if err == nil {
		t.Error("Collect() should return error")
	}
}

func TestAuditor_Collect_DescribeImagesError(t *testing.T) {
	now := time.Now()
	expectedErr := errors.New("API error")
	client := &fakeECRClient{
		repositories: []types.Repository{
			{
				RepositoryName: aws.String("repo1"),
				RepositoryUri:  aws.String("uri"),
				CreatedAt:      aws.Time(now),
			},
		},
		describeImagesErr: expectedErr,
	}

	auditor := NewAuditorWithClient(client, "us-east-1", "test")
	_, err := auditor.Collect(context.Background())

	if err == nil {
		t.Error("Collect() should return error when DescribeImages fails")
	}
}

func TestAuditor_Collect_ImageWithScanStatus(t *testing.T) {
	now := time.Now()
	client := &fakeECRClient{
		repositories: []types.Repository{
			{
				RepositoryName: aws.String("repo1"),
				RepositoryUri:  aws.String("uri"),
				CreatedAt:      aws.Time(now),
			},
		},
		images: map[string][]types.ImageDetail{
			"repo1": {
				{
					ImageDigest:      aws.String("sha256:abc"),
					ImagePushedAt:    aws.Time(now),
					ImageSizeInBytes: aws.Int64(1000),
					ImageScanStatus: &types.ImageScanStatus{
						Status: types.ScanStatusComplete,
					},
				},
			},
		},
	}

	auditor := NewAuditorWithClient(client, "us-east-1", "test")
	result, err := auditor.Collect(context.Background())

	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	if len(result.Repositories[0].Images) != 1 {
		t.Fatal("Expected 1 image")
	}

	if result.Repositories[0].Images[0].ImageScanStatus != "COMPLETE" {
		t.Errorf("ImageScanStatus = %s, expected COMPLETE", result.Repositories[0].Images[0].ImageScanStatus)
	}
}

func TestNewAuditorWithClient(t *testing.T) {
	client := &fakeECRClient{}
	auditor := NewAuditorWithClient(client, "us-west-2", "myprofile")

	if auditor.region != "us-west-2" {
		t.Errorf("region = %s, expected us-west-2", auditor.region)
	}
	if auditor.profile != "myprofile" {
		t.Errorf("profile = %s, expected myprofile", auditor.profile)
	}
}
