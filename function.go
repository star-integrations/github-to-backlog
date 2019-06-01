package github

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-yaml/yaml"
	"gopkg.in/go-playground/webhooks.v5/github"
)

type config struct {
	Backlog struct {
		APIKey     string `yaml:"apiKey"`
		ProjectKey string `yaml:"projectKey"`
		SpaceKey   string `yaml:"spaceKey"`
	} `yaml:"backlog"`
}

func loadConfig() *config {
	buf, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Printf("[loadConfig] Error: %#v", err)
		return nil
	}
	conf := &config{}
	err = yaml.Unmarshal(buf, conf)
	if err != nil {
		log.Printf("[loadConfig] Error: %#v", err)
		return nil
	}
	return conf
}

var conf *config

func Callback(w http.ResponseWriter, r *http.Request) {
	if conf == nil {
		conf = loadConfig()
	}

	branchesRe := regexp.MustCompile(conf.Backlog.ProjectKey + "-?[0-9]+")

	hook, _ := github.New()
	payload, err := hook.Parse(r, github.PushEvent, github.PullRequestEvent)
	if err != nil {
		if err == github.ErrEventNotFound {
			// ok event wasn;t one of the ones asked to be parsed
		} else {
			log.Printf("[hook.Parse] Error: %#v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
	}

	switch payload.(type) {
	case github.PushPayload:
		push := payload.(github.PushPayload)

		branches := strings.Replace(push.Ref, "refs/heads/", "", -1)
		issueMap := map[string]bool{}
		issueIDs := branchesRe.FindAllString(branches, -1)
		for _, issueID := range issueIDs {
			issueMap[issueID] = true
		}

		msg := ""
		action := ""
		if push.Created {
			action = "created"
		} else if push.Deleted {
			action = "deleted"
		}

		if action != "" {
			msg = fmt.Sprintf("%s %s\n", branches, action)
		}

		if len(push.Commits) > 0 {
			commit := push.Commits[len(push.Commits)-1]
			msg += fmt.Sprintf("\n[%s](%s) (%s)\n> %s\n> by %s(%s) <%s>\n",
				commit.ID,
				commit.URL,
				branches,
				strings.Replace(strings.TrimSpace(commit.Message), "\n", "\n> ", -1),
				commit.Committer.Username,
				commit.Committer.Name,
				commit.Committer.Email)
			issueIDs := branchesRe.FindAllString(commit.Message, -1)
			for _, issueID := range issueIDs {
				issueMap[issueID] = true
			}
		}

		for issueID := range issueMap {
			data := url.Values{}
			data.Add("content", msg)
			url := fmt.Sprintf("https://%s.backlog.jp/api/v2/issues/%s/comments?apiKey=%s", conf.Backlog.SpaceKey, issueID, conf.Backlog.APIKey)
			_, err := http.PostForm(url, data)
			if err != nil {
				log.Printf("[backlog] Error: %#v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))

	case github.PullRequestPayload:
		pr := payload.(github.PullRequestPayload)

		branches := strings.Replace(pr.PullRequest.Head.Ref, "refs/heads/", "", -1)
		issueMap := map[string]bool{}
		issueIDs := branchesRe.FindAllString(branches, -1)
		for _, issueID := range issueIDs {
			issueMap[issueID] = true
		}

		if pr.Action == "synchronize" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Success"))
			return
		}
		msg := fmt.Sprintf("PR %s %s", pr.PullRequest.HTMLURL, pr.Action)

		for issueID := range issueMap {
			data := url.Values{}
			data.Add("content", msg)
			apiURL := fmt.Sprintf("https://%s.backlog.jp/api/v2/issues/%s/comments?apiKey=%s", conf.Backlog.SpaceKey, issueID, conf.Backlog.APIKey)
			_, err := http.PostForm(apiURL, data)
			if err != nil {
				log.Printf("[backlog] Error: %#v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}

			data = url.Values{}
			data.Add("customField_69737", fmt.Sprintf("%d(%s)", pr.PullRequest.Number, pr.Action))
			apiURL = fmt.Sprintf("https://%s.backlog.jp/api/v2/issues/%s?apiKey=%s", conf.Backlog.SpaceKey, issueID, conf.Backlog.APIKey)
			req, err := http.NewRequest(http.MethodPatch, apiURL, strings.NewReader(data.Encode()))
			if err != nil {
				log.Printf("[backlog] Error: %#v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			_, err = http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("[backlog] Error: %#v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))

	default:
		msg := "Unkown payload"
		w.Write([]byte((msg)))
	}
}
