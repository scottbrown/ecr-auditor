package ecrauditor

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

type ECRClient interface {
	DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error)
	DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error)
	GetLifecyclePolicy(ctx context.Context, params *ecr.GetLifecyclePolicyInput, optFns ...func(*ecr.Options)) (*ecr.GetLifecyclePolicyOutput, error)
}

type Auditor struct {
	client  ECRClient
	region  string
	profile string
}

func NewAuditor(ctx context.Context, region, profile string) (*Auditor, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, err
	}

	client := ecr.NewFromConfig(cfg)
	return &Auditor{
		client:  client,
		region:  region,
		profile: profile,
	}, nil
}

func NewAuditorWithClient(client ECRClient, region, profile string) *Auditor {
	return &Auditor{
		client:  client,
		region:  region,
		profile: profile,
	}
}

func (a *Auditor) Collect(ctx context.Context) (*AuditResult, error) {
	result := &AuditResult{
		Region:      a.region,
		Profile:     a.profile,
		CollectedAt: time.Now(),
	}

	repos, err := a.listRepositories(ctx)
	if err != nil {
		return nil, err
	}

	result.RepositoryCount = len(repos)
	result.Repositories = repos

	return result, nil
}

func (a *Auditor) listRepositories(ctx context.Context) ([]Repository, error) {
	var repositories []Repository
	var nextToken *string

	for {
		output, err := a.client.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, repo := range output.Repositories {
			repository, err := a.processRepository(ctx, repo)
			if err != nil {
				return nil, err
			}
			repositories = append(repositories, repository)
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return repositories, nil
}

func (a *Auditor) processRepository(ctx context.Context, repo types.Repository) (Repository, error) {
	repository := Repository{
		Name:      aws.ToString(repo.RepositoryName),
		URI:       aws.ToString(repo.RepositoryUri),
		CreatedAt: aws.ToTime(repo.CreatedAt),
	}

	images, err := a.listImages(ctx, repository.Name)
	if err != nil {
		return Repository{}, err
	}
	repository.Images = images
	repository.ImageCount = len(images)

	policy, err := a.getLifecyclePolicy(ctx, repository.Name)
	if err == nil {
		repository.LifecyclePolicy = policy
	}

	return repository, nil
}

func (a *Auditor) listImages(ctx context.Context, repoName string) ([]Image, error) {
	var images []Image
	var nextToken *string

	for {
		output, err := a.client.DescribeImages(ctx, &ecr.DescribeImagesInput{
			RepositoryName: aws.String(repoName),
			NextToken:      nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, img := range output.ImageDetails {
			image := Image{
				Digest:    aws.ToString(img.ImageDigest),
				Tags:      img.ImageTags,
				PushedAt:  aws.ToTime(img.ImagePushedAt),
				SizeBytes: aws.ToInt64(img.ImageSizeInBytes),
			}
			if img.ImageScanStatus != nil {
				image.ImageScanStatus = string(img.ImageScanStatus.Status)
			}
			images = append(images, image)
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return images, nil
}

func (a *Auditor) getLifecyclePolicy(ctx context.Context, repoName string) (*LifecyclePolicy, error) {
	output, err := a.client.GetLifecyclePolicy(ctx, &ecr.GetLifecyclePolicyInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		return nil, err
	}

	return &LifecyclePolicy{
		RegistryID:     aws.ToString(output.RegistryId),
		RepositoryName: aws.ToString(output.RepositoryName),
		PolicyText:     aws.ToString(output.LifecyclePolicyText),
	}, nil
}
