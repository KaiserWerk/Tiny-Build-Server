package network

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
)

// CheckPayloadRequestHeader checks the existence and values taken from HTTP request headers
// from the given HTTP request
func CheckPayloadRequestHeader(content entity.BuildDefinitionContent, r *http.Request) error {
	var err error

	switch content.Repository.Hoster {
	case "bitbucket":
		headers := []string{"X-Event-Key", "X-Hook-Uuid", "X-Request-Uuid", "X-Attempt-Number"}
		for _, h := range headers {
			if _, err = helper.GetHeaderIfSet(r, h); err != nil {
				return fmt.Errorf("bitbucket: could not get header %s", h)
			}
		}

		var payload entity.BitBucketPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return fmt.Errorf("bitbucket: could not decode json payload: %s", err.Error())
		}
		_ = r.Body.Close()
		if payload.Push.Changes[0].New.Name != content.Repository.Branch {
			return fmt.Errorf("bitbucket: branch names do not match (from payload: %s, from build definition: %s)", payload.Push.Changes[0].New.Name, content.Repository.Branch)
		}
		if payload.Repository.FullName != content.Repository.Name {
			return fmt.Errorf("bitbucket: repository names do not match (from payload: %s, from build definition: %s)", payload.Repository.FullName, content.Repository.Name)
		}
	case "github":
		headers := []string{"X-GitHub-Delivery", "X-GitHub-Event", "X-Hub-Signature"}
		for _, h := range headers {
			if _, err = helper.GetHeaderIfSet(r, h); err != nil {
				return fmt.Errorf("github: could not get header %s", h)
			}
		}

		var payload entity.GitHubPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return fmt.Errorf("github: could not decode json payload")
		}
		_ = r.Body.Close()
		if payload.Repository.DefaultBranch != content.Repository.Branch {
			return fmt.Errorf("github: branch names do not match (from payload: %s, from build definition: %s)", payload.Repository.DefaultBranch, content.Repository.Branch)
		}
		if payload.Repository.FullName != content.Repository.Name {
			return fmt.Errorf("github: repository names do not match (from payload: %s, from build definition: %s)", payload.Repository.FullName, content.Repository.Name)
		}
	case "gitlab":
		headers := []string{"X-GitLab-Event"}
		for _, h := range headers {
			if _, err = helper.GetHeaderIfSet(r, h); err != nil {
				return fmt.Errorf("gitlab: could not get header %s", h)
			}
		}

		var payload entity.GitLabPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		_ = r.Body.Close()
		if err != nil {
			return fmt.Errorf("gitlab: could not decode json payload: %s", err.Error())
		}
		branch := strings.Split(payload.Ref, "/")[2]
		if branch != content.Repository.Branch {
			return fmt.Errorf("gitlab: branch names do not match (from payload: %s, from build definition: %s)", branch, content.Repository.Branch)
		}
		if payload.Project.PathWithNamespace != content.Repository.Name {
			return fmt.Errorf("gitlab: repository names do not match (from payload: %s, from build definition: %s)", payload.Project.PathWithNamespace, content.Repository.Name)
		}
	case "gitea":
		headers := []string{"X-Gitea-Delivery", "X-Gitea-Event"}
		for _, h := range headers {
			if _, err = helper.GetHeaderIfSet(r, h); err != nil {
				return fmt.Errorf("gitea: could not get header %s", h)
			}
		}

		var payload entity.GiteaPushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return fmt.Errorf("gitea: could not decode json payload: %s", err.Error())
		}
		_ = r.Body.Close()

		branch := strings.Split(payload.Ref, "/")[2]
		if branch != content.Repository.Branch {
			return fmt.Errorf("gitea: branch names do not match (from payload: %s, from build definition: %s)", branch, content.Repository.Branch)
		}
		if payload.Repository.FullName != content.Repository.Name {
			return fmt.Errorf("gitea: repository names do not match (from payload: %s, from build definition: %s)"+payload.Repository.FullName, content.Repository.Name)
		}
	case "azure_devops":
		headers := []string{"X-Request-Type"}
		for _, h := range headers {
			if _, err = helper.GetHeaderIfSet(r, h); err != nil {
				return fmt.Errorf("azure devops: could not get header %s", h)
			}
		}

		var payload entity.AzurePushPayload
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return fmt.Errorf("azure devops: could not decode json payload: %s", err.Error())
		}
		_ = r.Body.Close()
		// the name is supplied  in the form of "refs/heads/<branch>"
		if payload.Resource.RefUpdates[0].Name != "refs/heads/"+content.Repository.Branch {
			return fmt.Errorf("azure devops: branch names do not match (from payload: %s, from build definition: %s)",
				payload.Resource.RefUpdates[0].Name, "refs/heads/"+content.Repository.Branch)
		}
		if payload.Resource.Repository.Name != content.Repository.Name {
			return fmt.Errorf("azure devops: repository names do not match (from payload: %s, from build definition: %s)",
				payload.Resource.Repository.Name, content.Repository.Name)
		}
	default:
		return fmt.Errorf("unrecognized git hoster %s", content.Repository.Hoster)
	}

	return nil
}
