package data

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

type RepositoryData struct {
	Name          string
	NameWithOwner string
	Description   string
	UpdatedAt     time.Time
	CreatedAt     time.Time
	PushedAt      time.Time
	Url           string
	IsPrivate     bool
	IsArchived    bool
	IsFork        bool
	StargazerCount int
	ForkCount      int
	PrimaryLanguage struct {
		Name  string
		Color string
	}
	Owner struct {
		Login string
	}
}

func (data RepositoryData) GetTitle() string {
	if data.Description != "" {
		return data.Description
	}
	return data.Name
}

func (data RepositoryData) GetRepoNameWithOwner() string {
	return data.NameWithOwner
}

func (data RepositoryData) GetUrl() string {
	return data.Url
}

func (data RepositoryData) GetUpdatedAt() time.Time {
	return data.UpdatedAt
}

func (data RepositoryData) GetCreatedAt() time.Time {
	return data.CreatedAt
}

func (data RepositoryData) GetNumber() int {
	return 0
}

func makeRepositoriesQuery(query string) string {
	return fmt.Sprintf("%s sort:updated", query)
}

type RepositoriesResponse struct {
	Repositories []RepositoryData
	TotalCount   int
	PageInfo     PageInfo
}

func FetchRepositories(query string, limit int, pageInfo *PageInfo) (RepositoriesResponse, error) {
	var err error
	if client == nil {
		client, err = gh.DefaultGraphQLClient()
	}

	if err != nil {
		return RepositoriesResponse{}, err
	}

	var queryResult struct {
		Search struct {
			Nodes []struct {
				Repository RepositoryData `graphql:"... on Repository"`
			}
			RepositoryCount int
			PageInfo        PageInfo
		} `graphql:"search(type: REPOSITORY, first: $limit, after: $endCursor, query: $query)"`
	}
	var endCursor *string
	if pageInfo != nil {
		endCursor = &pageInfo.EndCursor
	}
	variables := map[string]any{
		"query":     graphql.String(makeRepositoriesQuery(query)),
		"limit":     graphql.Int(limit),
		"endCursor": (*graphql.String)(endCursor),
	}
	log.Debug("Fetching repositories", "query", query, "limit", limit, "endCursor", endCursor)
	err = client.Query("SearchRepositories", &queryResult, variables)
	if err != nil {
		return RepositoriesResponse{}, err
	}
	log.Debug("Successfully fetched repositories", "count", queryResult.Search.RepositoryCount)

	repositories := make([]RepositoryData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		if node.Repository.IsArchived {
			continue
		}
		repositories = append(repositories, node.Repository)
	}

	return RepositoriesResponse{
		Repositories: repositories,
		TotalCount:   queryResult.Search.RepositoryCount,
		PageInfo:     queryResult.Search.PageInfo,
	}, nil
}