package entity

import "time"

type GiteaPushPayload struct {
	Secret     string `json:"secret"`
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	CompareURL string `json:"compare_url"`
	Commits    []struct {
		ID      string `json:"Id"`
		Message string `json:"message"`
		URL     string `json:"url"`
		Author  struct {
			Name     string `json:"name"`
			Email    string `json:"Email"`
			Username string `json:"username"`
		} `json:"author"`
		Committer struct {
			Name     string `json:"name"`
			Email    string `json:"Email"`
			Username string `json:"username"`
		} `json:"committer"`
		Verification interface{}   `json:"verification"`
		Timestamp    time.Time     `json:"timestamp"`
		Added        []interface{} `json:"added"`
		Removed      []interface{} `json:"removed"`
		Modified     []string      `json:"modified"`
	} `json:"commits"`
	HeadCommit interface{} `json:"head_commit"`
	Repository struct {
		ID    int `json:"Id"`
		Owner struct {
			ID        int       `json:"Id"`
			Login     string    `json:"login"`
			FullName  string    `json:"full_name"`
			Email     string    `json:"Email"`
			AvatarURL string    `json:"avatar_url"`
			Language  string    `json:"language"`
			IsAdmin   bool      `json:"is_admin"`
			LastLogin time.Time `json:"last_login"`
			Created   time.Time `json:"created"`
			Username  string    `json:"username"`
		} `json:"owner"`
		Name            string      `json:"name"`
		FullName        string      `json:"full_name"`
		Description     string      `json:"description"`
		Empty           bool        `json:"empty"`
		Private         bool        `json:"private"`
		Fork            bool        `json:"fork"`
		Template        bool        `json:"template"`
		Parent          interface{} `json:"parent"`
		Mirror          bool        `json:"mirror"`
		Size            int         `json:"size"`
		HTMLURL         string      `json:"html_url"`
		SSHURL          string      `json:"ssh_url"`
		CloneURL        string      `json:"clone_url"`
		OriginalURL     string      `json:"original_url"`
		Website         string      `json:"website"`
		StarsCount      int         `json:"stars_count"`
		ForksCount      int         `json:"forks_count"`
		WatchersCount   int         `json:"watchers_count"`
		OpenIssuesCount int         `json:"open_issues_count"`
		OpenPrCounter   int         `json:"open_pr_counter"`
		ReleaseCounter  int         `json:"release_counter"`
		DefaultBranch   string      `json:"default_branch"`
		Archived        bool        `json:"archived"`
		CreatedAt       time.Time   `json:"created_at"`
		UpdatedAt       time.Time   `json:"updated_at"`
		Permissions     struct {
			Admin bool `json:"Admin"`
			Push  bool `json:"push"`
			Pull  bool `json:"pull"`
		} `json:"permissions"`
		HasIssues       bool `json:"has_issues"`
		InternalTracker struct {
			EnableTimeTracker                bool `json:"enable_time_tracker"`
			AllowOnlyContributorsToTrackTime bool `json:"allow_only_contributors_to_track_time"`
			EnableIssueDependencies          bool `json:"enable_issue_dependencies"`
		} `json:"internal_tracker"`
		HasWiki                   bool   `json:"has_wiki"`
		HasPullRequests           bool   `json:"has_pull_requests"`
		IgnoreWhitespaceConflicts bool   `json:"ignore_whitespace_conflicts"`
		AllowMergeCommits         bool   `json:"allow_merge_commits"`
		AllowRebase               bool   `json:"allow_rebase"`
		AllowRebaseExplicit       bool   `json:"allow_rebase_explicit"`
		AllowSquashMerge          bool   `json:"allow_squash_merge"`
		AvatarURL                 string `json:"avatar_url"`
	} `json:"repository"`
	Pusher struct {
		ID        int       `json:"Id"`
		Login     string    `json:"login"`
		FullName  string    `json:"full_name"`
		Email     string    `json:"Email"`
		AvatarURL string    `json:"avatar_url"`
		Language  string    `json:"language"`
		IsAdmin   bool      `json:"is_admin"`
		LastLogin time.Time `json:"last_login"`
		Created   time.Time `json:"created"`
		Username  string    `json:"username"`
	} `json:"pusher"`
	Sender struct {
		ID        int       `json:"Id"`
		Login     string    `json:"login"`
		FullName  string    `json:"full_name"`
		Email     string    `json:"Email"`
		AvatarURL string    `json:"avatar_url"`
		Language  string    `json:"language"`
		IsAdmin   bool      `json:"is_admin"`
		LastLogin time.Time `json:"last_login"`
		Created   time.Time `json:"created"`
		Username  string    `json:"username"`
	} `json:"sender"`
}
