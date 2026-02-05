package githubapi

import (
	"context"
	"strings"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/parsing"
	"github.com/google/go-github/v57/github"
	"github.com/samber/lo"
)

// FindWorkflows loads all the GitHub Actions workflows for a given repository.
// jwt is the JSON Web Token used for authentication.
// repo is the repository in the format "owner/repo".
func FindWorkflows(client *github.Client, repo string) []string {
	ctx := context.Background()

	owner, repoName, err := parsing.SplitRepo(repo)
	if err != nil {
		return []string{}
	}

	// List contents of .github/workflows directory
	_, dirContent, _, err := client.Repositories.GetContents(ctx, owner, repoName, ".github/workflows", nil)
	if err != nil {
		println("Error fetching workflows for repo", repo, ":", err.Error())
		return []string{}
	}

	files := lo.Filter(dirContent, func(item *github.RepositoryContent, index int) bool {
		return item.GetType() == "file" && strings.HasSuffix(strings.ToLower(item.GetName()), ".yml") || strings.HasSuffix(strings.ToLower(item.GetName()), ".yaml")
	})

	return lo.Map(files, func(item *github.RepositoryContent, index int) string {
		return item.GetName()
	})
}

func WorkflowToString(client *github.Client, repo string, workflow string) string {
	if client == nil {
		return ""
	}

	ctx := context.Background()

	owner, repoName, err := parsing.SplitRepo(repo)
	if err != nil {
		return ""
	}

	// Get file content
	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repoName, ".github/workflows/"+workflow, nil)
	if err != nil {
		return ""
	}

	// Decode content
	contentStr, err := fileContent.GetContent()
	if err != nil {
		return ""
	}

	return contentStr
}

func FindContributorsToWorkflow(client *github.Client, repo string, workflow string) []string {
	if client == nil {
		return []string{}
	}

	ctx := context.Background()

	// Split repo into owner and name
	owner, repoName, err := parsing.SplitRepo(repo)
	if err != nil {
		return []string{}
	}

	// Construct the workflow file path
	workflowPath := ".github/workflows/" + workflow

	// Get commits for the specific workflow file
	opts := &github.CommitsListOptions{
		Path: workflowPath,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// Track unique contributors
	var contributors []string

	// Fetch all commits for the workflow file (handle pagination)
	for {
		commits, resp, err := client.Repositories.ListCommits(ctx, owner, repoName, opts)
		if err != nil {
			return []string{}
		}

		// Extract unique contributor names
		authors := lo.Filter(commits, func(item *github.RepositoryCommit, index int) bool {
			return item.Commit != nil && item.Commit.Author != nil && item.Commit.Author.Name != nil
		})

		authorNames := lo.Map(authors, func(item *github.RepositoryCommit, index int) string {
			return *item.Commit.Author.Name
		})

		contributors = lo.Uniq(append(contributors, authorNames...))

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return contributors
}

func GetWorkflowAdvisories(client *github.Client, repo string) []string {
	if client == nil {
		return []string{}
	}

	ctx := context.Background()

	owner, repoName, err := parsing.SplitRepo(repo)
	if err != nil {
		return []string{}
	}

	opts := &github.ListRepositorySecurityAdvisoriesOptions{
		ListCursorOptions: github.ListCursorOptions{
			PerPage: 100,
		},
	}

	var advisories []string

	// Fetch all security advisories for the repository (handle pagination)
	for {
		advisoryList, resp, err := client.SecurityAdvisories.ListRepositorySecurityAdvisories(ctx, owner, repoName, opts)
		if err != nil {
			// If there's an error (e.g., no access, repo doesn't exist), return empty list
			return []string{}
		}

		// Extract advisory IDs or summaries
		for _, advisory := range advisoryList {
			if advisory.GHSAID != nil {
				advisories = append(advisories, *advisory.GHSAID)
			}
		}

		// Check if there are more pages
		if resp.After == "" {
			break
		}
		opts.After = resp.After
	}

	if advisories == nil {
		return []string{}
	}

	return advisories
}
