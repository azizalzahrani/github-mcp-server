package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListDiscussions creates a tool to list discussions in a GitHub repository
func ListDiscussions(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_discussions",
			mcp.WithDescription(t("TOOL_LIST_DISCUSSIONS_DESCRIPTION", "List discussions in a GitHub repository with filtering options")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("direction",
				mcp.Description("Sort direction ('asc', 'desc')"),
				mcp.Enum("asc", "desc"),
			),
			mcp.WithString("category_id",
				mcp.Description("Filter by category ID"),
			),
			mcp.WithString("pinned",
				mcp.Description("Filter by pinned status ('true', 'false')"),
				mcp.Enum("true", "false"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.DiscussionListOptions{}

			// Set optional parameters if provided
			direction, err := OptionalParam[string](request, "direction")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if direction != "" {
				opts.Direction = direction
			}

			categoryID, err := OptionalParam[string](request, "category_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if categoryID != "" {
				opts.CategoryID = categoryID
			}

			pinnedStr, err := OptionalParam[string](request, "pinned")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if pinnedStr == "true" {
				pinned := true
				opts.Pinned = &pinned
			} else if pinnedStr == "false" {
				pinned := false
				opts.Pinned = &pinned
			}

			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			opts.ListOptions = github.ListOptions{
				Page:    pagination.page,
				PerPage: pagination.perPage,
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			discussions, resp, err := client.Discussions.ListDiscussions(ctx, owner, repo, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list discussions: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to list discussions: %s", string(body))), nil
			}

			r, err := json.Marshal(discussions)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal discussions: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// GetDiscussion creates a tool to get details of a specific discussion in a GitHub repository
func GetDiscussion(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_discussion",
			mcp.WithDescription(t("TOOL_GET_DISCUSSION_DESCRIPTION", "Get details of a specific discussion in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("The owner of the repository"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("The name of the repository"),
			),
			mcp.WithNumber("discussion_number",
				mcp.Required(),
				mcp.Description("The number of the discussion"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			discussionNumber, err := RequiredInt(request, "discussion_number")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			discussion, resp, err := client.Discussions.GetDiscussion(ctx, owner, repo, discussionNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to get discussion: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get discussion: %s", string(body))), nil
			}

			r, err := json.Marshal(discussion)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal discussion: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// GetDiscussionCategories creates a tool to get discussion categories in a GitHub repository
func GetDiscussionCategories(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_discussion_categories",
			mcp.WithDescription(t("TOOL_GET_DISCUSSION_CATEGORIES_DESCRIPTION", "Get discussion categories in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			opts := &github.ListOptions{
				Page:    pagination.page,
				PerPage: pagination.perPage,
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			categories, resp, err := client.Discussions.ListDiscussionCategories(ctx, owner, repo, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to get discussion categories: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get discussion categories: %s", string(body))), nil
			}

			r, err := json.Marshal(categories)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal categories: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// GetDiscussionComments creates a tool to get comments for a GitHub discussion
func GetDiscussionComments(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_discussion_comments",
			mcp.WithDescription(t("TOOL_GET_DISCUSSION_COMMENTS_DESCRIPTION", "Get comments for a GitHub discussion")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithNumber("discussion_number",
				mcp.Required(),
				mcp.Description("Discussion number"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			discussionNumber, err := RequiredInt(request, "discussion_number")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			opts := &github.DiscussionCommentListOptions{
				ListOptions: github.ListOptions{
					Page:    pagination.page,
					PerPage: pagination.perPage,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			comments, resp, err := client.Discussions.ListDiscussionComments(ctx, owner, repo, discussionNumber, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to get discussion comments: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get discussion comments: %s", string(body))), nil
			}

			r, err := json.Marshal(comments)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal comments: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// AddDiscussionComment creates a tool to add a comment to a discussion
func AddDiscussionComment(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("add_discussion_comment",
			mcp.WithDescription(t("TOOL_ADD_DISCUSSION_COMMENT_DESCRIPTION", "Add a comment to an existing discussion")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithNumber("discussion_number",
				mcp.Required(),
				mcp.Description("Discussion number to comment on"),
			),
			mcp.WithString("body",
				mcp.Required(),
				mcp.Description("Comment text"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			discussionNumber, err := RequiredInt(request, "discussion_number")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body, err := requiredParam[string](request, "body")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			comment := &github.DiscussionComment{
				Body: github.Ptr(body),
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			createdComment, resp, err := client.Discussions.CreateDiscussionComment(ctx, owner, repo, discussionNumber, comment)
			if err != nil {
				return nil, fmt.Errorf("failed to create discussion comment: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to create discussion comment: %s", string(body))), nil
			}

			r, err := json.Marshal(createdComment)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// CreateDiscussion creates a tool to create a new discussion in a GitHub repository
func CreateDiscussion(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_discussion",
			mcp.WithDescription(t("TOOL_CREATE_DISCUSSION_DESCRIPTION", "Create a new discussion in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("title",
				mcp.Required(),
				mcp.Description("Discussion title"),
			),
			mcp.WithString("body",
				mcp.Required(),
				mcp.Description("Discussion body content"),
			),
			mcp.WithString("category_id",
				mcp.Required(),
				mcp.Description("Category ID for the discussion"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			title, err := requiredParam[string](request, "title")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body, err := requiredParam[string](request, "body")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			categoryID, err := requiredParam[string](request, "category_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			discussionRequest := &github.DiscussionRequest{
				Title:      github.Ptr(title),
				Body:       github.Ptr(body),
				CategoryID: github.Ptr(categoryID),
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			discussion, resp, err := client.Discussions.CreateDiscussion(ctx, owner, repo, discussionRequest)
			if err != nil {
				return nil, fmt.Errorf("failed to create discussion: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to create discussion: %s", string(body))), nil
			}

			r, err := json.Marshal(discussion)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}
