package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ListDiscussions(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ListDiscussions(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "list_discussions", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "direction")
	assert.Contains(t, tool.InputSchema.Properties, "category_id")
	assert.Contains(t, tool.InputSchema.Properties, "pinned")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})

	// Setup mock discussions for success case
	mockDiscussions := []*github.Discussion{
		{
			Number:      github.Ptr(123),
			Title:       github.Ptr("First Discussion"),
			Body:        github.Ptr("This is the first test discussion"),
			HTMLURL:     github.Ptr("https://github.com/owner/repo/discussions/123"),
			CreatedAt:   &github.Timestamp{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
			CategoryID:  github.Ptr("1"),
			Category:    &github.DiscussionCategory{ID: github.Ptr("1"), Name: github.Ptr("General")},
			AnswerHTMLURL: github.Ptr("https://github.com/owner/repo/discussions/123#discussioncomment-1234"),
		},
		{
			Number:      github.Ptr(456),
			Title:       github.Ptr("Second Discussion"),
			Body:        github.Ptr("This is the second test discussion"),
			HTMLURL:     github.Ptr("https://github.com/owner/repo/discussions/456"),
			CreatedAt:   &github.Timestamp{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)},
			CategoryID:  github.Ptr("2"),
			Category:    &github.DiscussionCategory{ID: github.Ptr("2"), Name: github.Ptr("Q&A")},
		},
	}

	tests := []struct {
		name              string
		mockedClient      *http.Client
		requestArgs       map[string]interface{}
		expectError       bool
		expectedErrMsg    string
		expectedDiscussions []*github.Discussion
	}{
		{
			name: "list discussions with minimal parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposDiscussionsByOwnerByRepo,
					mockDiscussions,
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError:          false,
			expectedDiscussions: mockDiscussions,
		},
		{
			name: "list discussions with all parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposDiscussionsByOwnerByRepo,
					expectQueryParams(t, map[string]string{
						"direction":  "desc",
						"category":   "1",
						"pinned":     "true",
						"page":       "1",
						"per_page":   "30",
					}).andThen(
						mockResponse(t, http.StatusOK, mockDiscussions),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":       "owner",
				"repo":        "repo",
				"direction":   "desc",
				"category_id": "1",
				"pinned":      "true",
				"page":        float64(1),
				"perPage":     float64(30),
			},
			expectError:          false,
			expectedDiscussions: mockDiscussions,
		},
		{
			name: "list discussions fails with error",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposDiscussionsByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Repository not found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "nonexistent",
				"repo":  "repo",
			},
			expectError:    true,
			expectedErrMsg: "failed to list discussions",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := ListDiscussions(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				} else {
					// For errors returned as part of the result, not as an error
					assert.NotNil(t, result)
					textContent := getTextResult(t, result)
					assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				}
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedDiscussions []*github.Discussion
			err = json.Unmarshal([]byte(textContent.Text), &returnedDiscussions)
			require.NoError(t, err)

			assert.Len(t, returnedDiscussions, len(tc.expectedDiscussions))
			for i, discussion := range returnedDiscussions {
				assert.Equal(t, *tc.expectedDiscussions[i].Number, *discussion.Number)
				assert.Equal(t, *tc.expectedDiscussions[i].Title, *discussion.Title)
				assert.Equal(t, *tc.expectedDiscussions[i].HTMLURL, *discussion.HTMLURL)
				assert.Equal(t, *tc.expectedDiscussions[i].CategoryID, *discussion.CategoryID)
			}
		})
	}
}

func Test_GetDiscussion(t *testing.T) {
	// Verify tool definition
	mockClient := github.NewClient(nil)
	tool, _ := GetDiscussion(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "get_discussion", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "discussion_number")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "discussion_number"})

	// Setup mock discussion response
	mockDiscussion := &github.Discussion{
		Number:     github.Ptr(42),
		Title:      github.Ptr("Test Discussion"),
		Body:       github.Ptr("This is a test discussion"),
		HTMLURL:    github.Ptr("https://github.com/owner/repo/discussions/42"),
		CategoryID: github.Ptr("1"),
		Category:   &github.DiscussionCategory{ID: github.Ptr("1"), Name: github.Ptr("General")},
	}

	tests := []struct {
		name             string
		mockedClient     *http.Client
		requestArgs      map[string]interface{}
		expectError      bool
		expectedDiscussion *github.Discussion
		expectedErrMsg   string
	}{
		{
			name: "successful discussion retrieval",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposDiscussionsByOwnerByRepoByDiscussionNumber,
					mockDiscussion,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":             "owner",
				"repo":              "repo",
				"discussion_number": float64(42),
			},
			expectError:         false,
			expectedDiscussion: mockDiscussion,
		},
		{
			name: "discussion not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposDiscussionsByOwnerByRepoByDiscussionNumber,
					mockResponse(t, http.StatusNotFound, `{"message": "Discussion not found"}`),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":             "owner",
				"repo":              "repo",
				"discussion_number": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get discussion",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetDiscussion(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedDiscussion github.Discussion
			err = json.Unmarshal([]byte(textContent.Text), &returnedDiscussion)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedDiscussion.Number, *returnedDiscussion.Number)
			assert.Equal(t, *tc.expectedDiscussion.Title, *returnedDiscussion.Title)
			assert.Equal(t, *tc.expectedDiscussion.Body, *returnedDiscussion.Body)
			assert.Equal(t, *tc.expectedDiscussion.HTMLURL, *returnedDiscussion.HTMLURL)
			assert.Equal(t, *tc.expectedDiscussion.CategoryID, *returnedDiscussion.CategoryID)
		})
	}
}

func Test_GetDiscussionCategories(t *testing.T) {
	// Verify tool definition
	mockClient := github.NewClient(nil)
	tool, _ := GetDiscussionCategories(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "get_discussion_categories", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})

	// Setup mock categories
	mockCategories := []*github.DiscussionCategory{
		{
			ID:          github.Ptr("1"),
			Name:        github.Ptr("General"),
			Description: github.Ptr("General discussions"),
			Emoji:       github.Ptr("💬"),
		},
		{
			ID:          github.Ptr("2"),
			Name:        github.Ptr("Q&A"),
			Description: github.Ptr("Questions and answers"),
			Emoji:       github.Ptr("❓"),
		},
	}

	tests := []struct {
		name              string
		mockedClient      *http.Client
		requestArgs       map[string]interface{}
		expectError       bool
		expectedCategories []*github.DiscussionCategory
		expectedErrMsg    string
	}{
		{
			name: "get categories successful",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposDiscussionsCategoriesByOwnerByRepo,
					mockCategories,
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError:          false,
			expectedCategories: mockCategories,
		},
		{
			name: "get categories with pagination",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposDiscussionsCategoriesByOwnerByRepo,
					expectQueryParams(t, map[string]string{
						"page":     "2",
						"per_page": "10",
					}).andThen(
						mockResponse(t, http.StatusOK, mockCategories),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":   "owner",
				"repo":    "repo",
				"page":    float64(2),
				"perPage": float64(10),
			},
			expectError:          false,
			expectedCategories: mockCategories,
		},
		{
			name: "repository not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposDiscussionsCategoriesByOwnerByRepo,
					mockResponse(t, http.StatusNotFound, `{"message": "Repository not found"}`),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "nonexistent",
				"repo":  "repo",
			},
			expectError:    true,
			expectedErrMsg: "failed to get discussion categories",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetDiscussionCategories(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedCategories []*github.DiscussionCategory
			err = json.Unmarshal([]byte(textContent.Text), &returnedCategories)
			require.NoError(t, err)
			assert.Len(t, returnedCategories, len(tc.expectedCategories))
			for i, category := range returnedCategories {
				assert.Equal(t, *tc.expectedCategories[i].ID, *category.ID)
				assert.Equal(t, *tc.expectedCategories[i].Name, *category.Name)
				assert.Equal(t, *tc.expectedCategories[i].Description, *category.Description)
				assert.Equal(t, *tc.expectedCategories[i].Emoji, *category.Emoji)
			}
		})
	}
}

func Test_GetDiscussionComments(t *testing.T) {
	// Verify tool definition
	mockClient := github.NewClient(nil)
	tool, _ := GetDiscussionComments(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "get_discussion_comments", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "discussion_number")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "discussion_number"})

	// Setup mock comments
	mockComments := []*github.DiscussionComment{
		{
			ID:        github.Ptr(int64(123)),
			Number:    github.Ptr(1),
			Body:      github.Ptr("This is the first comment"),
			User:      &github.User{Login: github.Ptr("user1")},
			CreatedAt: &github.Timestamp{Time: time.Now().Add(-time.Hour * 24)},
			HTMLURL:   github.Ptr("https://github.com/owner/repo/discussions/42#discussioncomment-123"),
		},
		{
			ID:        github.Ptr(int64(456)),
			Number:    github.Ptr(2),
			Body:      github.Ptr("This is the second comment"),
			User:      &github.User{Login: github.Ptr("user2")},
			CreatedAt: &github.Timestamp{Time: time.Now().Add(-time.Hour)},
			HTMLURL:   github.Ptr("https://github.com/owner/repo/discussions/42#discussioncomment-456"),
		},
	}

	tests := []struct {
		name             string
		mockedClient     *http.Client
		requestArgs      map[string]interface{}
		expectError      bool
		expectedComments []*github.DiscussionComment
		expectedErrMsg   string
	}{
		{
			name: "successful comments retrieval",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposDiscussionsCommentsByOwnerByRepoByDiscussionNumber,
					mockComments,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":             "owner",
				"repo":              "repo",
				"discussion_number": float64(42),
			},
			expectError:       false,
			expectedComments: mockComments,
		},
		{
			name: "successful comments retrieval with pagination",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposDiscussionsCommentsByOwnerByRepoByDiscussionNumber,
					expectQueryParams(t, map[string]string{
						"page":     "2",
						"per_page": "10",
					}).andThen(
						mockResponse(t, http.StatusOK, mockComments),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":             "owner",
				"repo":              "repo",
				"discussion_number": float64(42),
				"page":              float64(2),
				"perPage":           float64(10),
			},
			expectError:       false,
			expectedComments: mockComments,
		},
		{
			name: "discussion not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposDiscussionsCommentsByOwnerByRepoByDiscussionNumber,
					mockResponse(t, http.StatusNotFound, `{"message": "Discussion not found"}`),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":             "owner",
				"repo":              "repo",
				"discussion_number": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get discussion comments",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetDiscussionComments(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedComments []*github.DiscussionComment
			err = json.Unmarshal([]byte(textContent.Text), &returnedComments)
			require.NoError(t, err)
			assert.Len(t, returnedComments, len(tc.expectedComments))
			for i, comment := range returnedComments {
				assert.Equal(t, *tc.expectedComments[i].Number, *comment.Number)
				assert.Equal(t, *tc.expectedComments[i].Body, *comment.Body)
				assert.Equal(t, *tc.expectedComments[i].User.Login, *comment.User.Login)
				assert.Equal(t, *tc.expectedComments[i].HTMLURL, *comment.HTMLURL)
			}
		})
	}
}

func Test_AddDiscussionComment(t *testing.T) {
	// Verify tool definition
	mockClient := github.NewClient(nil)
	tool, _ := AddDiscussionComment(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "add_discussion_comment", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "discussion_number")
	assert.Contains(t, tool.InputSchema.Properties, "body")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "discussion_number", "body"})

	// Setup mock comment for success case
	mockComment := &github.DiscussionComment{
		ID:     github.Ptr(int64(123)),
		Number: github.Ptr(1),
		Body:   github.Ptr("This is a test comment"),
		User:   &github.User{Login: github.Ptr("testuser")},
		HTMLURL: github.Ptr("https://github.com/owner/repo/discussions/42#discussioncomment-123"),
	}

	tests := []struct {
		name            string
		mockedClient    *http.Client
		requestArgs     map[string]interface{}
		expectError     bool
		expectedComment *github.DiscussionComment
		expectedErrMsg  string
	}{
		{
			name: "successful comment creation",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposDiscussionsCommentsByOwnerByRepoByDiscussionNumber,
					expectRequestBody(t, map[string]any{
						"body": "This is a test comment",
					}).andThen(
						mockResponse(t, http.StatusCreated, mockComment),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":             "owner",
				"repo":              "repo",
				"discussion_number": float64(42),
				"body":              "This is a test comment",
			},
			expectError:     false,
			expectedComment: mockComment,
		},
		{
			name: "comment creation fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposDiscussionsCommentsByOwnerByRepoByDiscussionNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Invalid request"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":             "owner",
				"repo":              "repo",
				"discussion_number": float64(42),
				"body":              "",
			},
			expectError:    false,
			expectedErrMsg: "missing required parameter: body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := AddDiscussionComment(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			if tc.expectedErrMsg != "" {
				require.NotNil(t, result)
				textContent := getTextResult(t, result)
				assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedComment github.DiscussionComment
			err = json.Unmarshal([]byte(textContent.Text), &returnedComment)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedComment.ID, *returnedComment.ID)
			assert.Equal(t, *tc.expectedComment.Body, *returnedComment.Body)
			assert.Equal(t, *tc.expectedComment.User.Login, *returnedComment.User.Login)
			assert.Equal(t, *tc.expectedComment.HTMLURL, *returnedComment.HTMLURL)
		})
	}
}

func Test_CreateDiscussion(t *testing.T) {
	// Verify tool definition
	mockClient := github.NewClient(nil)
	tool, _ := CreateDiscussion(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "create_discussion", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "title")
	assert.Contains(t, tool.InputSchema.Properties, "body")
	assert.Contains(t, tool.InputSchema.Properties, "category_id")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "title", "body", "category_id"})

	// Setup mock discussion for success case
	mockDiscussion := &github.Discussion{
		Number:     github.Ptr(123),
		Title:      github.Ptr("Test Discussion"),
		Body:       github.Ptr("This is a test discussion"),
		HTMLURL:    github.Ptr("https://github.com/owner/repo/discussions/123"),
		CategoryID: github.Ptr("1"),
		Category:   &github.DiscussionCategory{ID: github.Ptr("1"), Name: github.Ptr("General")},
	}

	tests := []struct {
		name              string
		mockedClient      *http.Client
		requestArgs       map[string]interface{}
		expectError       bool
		expectedDiscussion *github.Discussion
		expectedErrMsg    string
	}{
		{
			name: "successful discussion creation",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposDiscussionsByOwnerByRepo,
					expectRequestBody(t, map[string]any{
						"title":       "Test Discussion",
						"body":        "This is a test discussion",
						"category_id": "1",
					}).andThen(
						mockResponse(t, http.StatusCreated, mockDiscussion),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":       "owner",
				"repo":        "repo",
				"title":       "Test Discussion",
				"body":        "This is a test discussion",
				"category_id": "1",
			},
			expectError:         false,
			expectedDiscussion: mockDiscussion,
		},
		{
			name: "discussion creation fails with invalid category",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposDiscussionsByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Invalid category ID"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":       "owner",
				"repo":        "repo",
				"title":       "Test Discussion",
				"body":        "This is a test discussion",
				"category_id": "invalid",
			},
			expectError:    true,
			expectedErrMsg: "failed to create discussion",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := CreateDiscussion(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				} else {
					require.NotNil(t, result)
					textContent := getTextResult(t, result)
					assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				}
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedDiscussion github.Discussion
			err = json.Unmarshal([]byte(textContent.Text), &returnedDiscussion)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedDiscussion.Number, *returnedDiscussion.Number)
			assert.Equal(t, *tc.expectedDiscussion.Title, *returnedDiscussion.Title)
			assert.Equal(t, *tc.expectedDiscussion.Body, *returnedDiscussion.Body)
			assert.Equal(t, *tc.expectedDiscussion.HTMLURL, *returnedDiscussion.HTMLURL)
			assert.Equal(t, *tc.expectedDiscussion.CategoryID, *returnedDiscussion.CategoryID)
		})
	}
}
