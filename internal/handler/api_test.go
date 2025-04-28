package handler

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/sirupsen/logrus"
)

type DBServiceMock struct {
	AutoMigrateFunc                  func() error
	QuitFunc                         func()
	RowExistsFunc                    func(query string, args ...interface{}) bool
	FindBuildDefinitionFunc          func(cond string, args ...interface{}) (entity.BuildDefinition, error)
	GetAvailableVariablesForUserFunc func(userId uint) ([]entity.UserVariable, error)
}

func (m *DBServiceMock) AutoMigrate() error {

	return nil
}

func (m *DBServiceMock) Quit() {

}

func (m *DBServiceMock) RowExists(query string, args ...interface{}) bool {
	return true
}

func (m *DBServiceMock) FindBuildDefinition(cond string, args ...interface{}) (entity.BuildDefinition, error) {
	return entity.BuildDefinition{
		Token: "123abc",
	}, nil
}

func (m *DBServiceMock) GetAvailableVariablesForUser(userId uint) ([]entity.UserVariable, error) {
	return nil, nil
}

var mockPayload = `{
  "ref": "refs/heads/master",
  "before": "0311ed5c6bd5d9ed16f7520f62e04c051d748090",
  "after": "1316f73b181936990972ab07e3d6c215367bf8cc",
  "repository": {
    "id": 267527589,
    "node_id": "MDEwOlJlcG9zaXRvcnkyNjc1Mjc1ODk=",
    "name": "GitHub-Public-Golang-Test-Repo",
    "full_name": "KaiserWerk/GitHub-Public-Golang-Test-Repo",
    "private": false,
    "owner": {
      "name": "KaiserWerk",
      "email": "web@r-k.mx",
      "login": "KaiserWerk",
      "id": 12614975,
      "node_id": "MDQ6VXNlcjEyNjE0OTc1",
      "avatar_url": "https://avatars.githubusercontent.com/u/12614975?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/KaiserWerk",
      "html_url": "https://github.com/KaiserWerk",
      "followers_url": "https://api.github.com/users/KaiserWerk/followers",
      "following_url": "https://api.github.com/users/KaiserWerk/following{/other_user}",
      "gists_url": "https://api.github.com/users/KaiserWerk/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/KaiserWerk/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/KaiserWerk/subscriptions",
      "organizations_url": "https://api.github.com/users/KaiserWerk/orgs",
      "repos_url": "https://api.github.com/users/KaiserWerk/repos",
      "events_url": "https://api.github.com/users/KaiserWerk/events{/privacy}",
      "received_events_url": "https://api.github.com/users/KaiserWerk/received_events",
      "type": "User",
      "site_admin": false
    },
    "html_url": "https://github.com/KaiserWerk/GitHub-Public-Golang-Test-Repo",
    "description": null,
    "fork": false,
    "url": "https://github.com/KaiserWerk/GitHub-Public-Golang-Test-Repo",
    "forks_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/forks",
    "keys_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/keys{/key_id}",
    "collaborators_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/collaborators{/collaborator}",
    "teams_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/teams",
    "hooks_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/hooks",
    "issue_events_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/issues/events{/number}",
    "events_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/events",
    "assignees_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/assignees{/user}",
    "branches_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/branches{/branch}",
    "tags_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/tags",
    "blobs_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/git/blobs{/sha}",
    "git_tags_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/git/tags{/sha}",
    "git_refs_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/git/refs{/sha}",
    "trees_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/git/trees{/sha}",
    "statuses_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/statuses/{sha}",
    "languages_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/languages",
    "stargazers_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/stargazers",
    "contributors_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/contributors",
    "subscribers_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/subscribers",
    "subscription_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/subscription",
    "commits_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/commits{/sha}",
    "git_commits_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/git/commits{/sha}",
    "comments_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/comments{/number}",
    "issue_comment_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/issues/comments{/number}",
    "contents_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/contents/{+path}",
    "compare_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/compare/{base}...{head}",
    "merges_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/merges",
    "archive_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/{archive_format}{/ref}",
    "downloads_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/downloads",
    "issues_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/issues{/number}",
    "pulls_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/pulls{/number}",
    "milestones_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/milestones{/number}",
    "notifications_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/notifications{?since,all,participating}",
    "labels_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/labels{/name}",
    "releases_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/releases{/id}",
    "deployments_url": "https://api.github.com/repos/KaiserWerk/GitHub-Public-Golang-Test-Repo/deployments",
    "created_at": 1590652316,
    "updated_at": "2020-05-28T09:07:35Z",
    "pushed_at": 1619367026,
    "git_url": "git://github.com/KaiserWerk/GitHub-Public-Golang-Test-Repo.git",
    "ssh_url": "git@github.com:KaiserWerk/GitHub-Public-Golang-Test-Repo.git",
    "clone_url": "https://github.com/KaiserWerk/GitHub-Public-Golang-Test-Repo.git",
    "svn_url": "https://github.com/KaiserWerk/GitHub-Public-Golang-Test-Repo",
    "homepage": null,
    "size": 2,
    "stargazers_count": 0,
    "watchers_count": 0,
    "language": "Go",
    "has_issues": true,
    "has_projects": true,
    "has_downloads": true,
    "has_wiki": true,
    "has_pages": false,
    "forks_count": 0,
    "mirror_url": null,
    "archived": false,
    "disabled": false,
    "open_issues_count": 0,
    "license": null,
    "forks": 0,
    "open_issues": 0,
    "watchers": 0,
    "default_branch": "master",
    "stargazers": 0,
    "master_branch": "master"
  },
  "pusher": {
    "name": "KaiserWerk",
    "email": "web@r-k.mx"
  },
  "sender": {
    "login": "KaiserWerk",
    "id": 12614975,
    "node_id": "MDQ6VXNlcjEyNjE0OTc1",
    "avatar_url": "https://avatars.githubusercontent.com/u/12614975?v=4",
    "gravatar_id": "",
    "url": "https://api.github.com/users/KaiserWerk",
    "html_url": "https://github.com/KaiserWerk",
    "followers_url": "https://api.github.com/users/KaiserWerk/followers",
    "following_url": "https://api.github.com/users/KaiserWerk/following{/other_user}",
    "gists_url": "https://api.github.com/users/KaiserWerk/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/KaiserWerk/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/KaiserWerk/subscriptions",
    "organizations_url": "https://api.github.com/users/KaiserWerk/orgs",
    "repos_url": "https://api.github.com/users/KaiserWerk/repos",
    "events_url": "https://api.github.com/users/KaiserWerk/events{/privacy}",
    "received_events_url": "https://api.github.com/users/KaiserWerk/received_events",
    "type": "User",
    "site_admin": false
  },
  "created": false,
  "deleted": false,
  "forced": false,
  "base_ref": null,
  "compare": "https://github.com/KaiserWerk/GitHub-Public-Golang-Test-Repo/compare/0311ed5c6bd5...1316f73b1819",
  "commits": [
    {
      "id": "1316f73b181936990972ab07e3d6c215367bf8cc",
      "tree_id": "0153f36a0c8d03c4b95deff88bb86b9745c104ca",
      "distinct": true,
      "message": "Update test.txt",
      "timestamp": "2021-04-25T18:10:26+02:00",
      "url": "https://github.com/KaiserWerk/GitHub-Public-Golang-Test-Repo/commit/1316f73b181936990972ab07e3d6c215367bf8cc",
      "author": {
        "name": "Robin Kaiser",
        "email": "m@r-k.mx",
        "username": "KaiserWerk"
      },
      "committer": {
        "name": "GitHub",
        "email": "noreply@github.com",
        "username": "web-flow"
      },
      "added": [

      ],
      "removed": [

      ],
      "modified": [
        "test.txt"
      ]
    }
  ],
  "head_commit": {
    "id": "1316f73b181936990972ab07e3d6c215367bf8cc",
    "tree_id": "0153f36a0c8d03c4b95deff88bb86b9745c104ca",
    "distinct": true,
    "message": "Update test.txt",
    "timestamp": "2021-04-25T18:10:26+02:00",
    "url": "https://github.com/KaiserWerk/GitHub-Public-Golang-Test-Repo/commit/1316f73b181936990972ab07e3d6c215367bf8cc",
    "author": {
      "name": "Robin Kaiser",
      "email": "m@r-k.mx",
      "username": "KaiserWerk"
    },
    "committer": {
      "name": "GitHub",
      "email": "noreply@github.com",
      "username": "web-flow"
    },
    "added": [

    ],
    "removed": [

    ],
    "modified": [
      "test.txt"
    ]
  }
}`

func TestPayloadReceiveHandler(t *testing.T) {
	logger := logrus.New()
	logger.Out = io.Discard

	dbMock := &DBServiceMock{}

	handler := &HTTPHandler{
		Logger:    logrus.NewEntry(logger),
		DBService: dbMock,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/payload/receive?token=123abc", strings.NewReader(mockPayload))

	handler.PayloadReceiveHandler(w, r)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)

	}
}
