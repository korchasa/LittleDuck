package main

import (
    "bytes"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "github.com/davecgh/go-spew/spew"
    log "github.com/sirupsen/logrus"
    "io"
    "net/http"
    "net/url"
    "text/template"
)

type CommandJiraGetTasks struct {
    host  string
    user  string
    token string
}

func NewCommandJiraGetTasks() *CommandJiraGetTasks {
    return &CommandJiraGetTasks{
        host:  ensureEnv("JIRA_HOST"),
        user:  ensureEnv("JIRA_USER"),
        token: ensureEnv("JIRA_TOKEN"),
    }
}

func (c *CommandJiraGetTasks) Name() string {
    return "jiraSearchTasks"
}

func (c *CommandJiraGetTasks) Prompts() []PromptExample {
    return []PromptExample{
        {
            "покажи мои открытые задачи",
            CommandSpec{
                "jiraSearchTasks",
                []string{"assignee = currentUser() AND resolution = Unresolved ORDER BY updated DESC"},
            },
        },
        {
            "покажи мои задачи",
            CommandSpec{
                "jiraSearchTasks",
                []string{"assignee = currentUser() AND resolution = Unresolved ORDER BY updated DESC"},
            },
        },
        {
            "покажи все задачи, которые я создал",
            CommandSpec{
                "jiraSearchTasks",
                []string{"reporter = currentUser()"},
            },
        },
        {
            "покажи все недавно созданные задачи",
            CommandSpec{
                "jiraSearchTasks",
                []string{"created >= -1w order by created DESC"},
            },
        },
        {
            "покажи мои эпики",
            CommandSpec{
                "jiraSearchTasks",
                []string{"issuetype = Epic AND resolution = Unresolved AND (watcher = currentUser() OR assignee = currentUser()) order by created DESC"},
            },
        },
    }
}

func (c *CommandJiraGetTasks) Execute(args []string) (string, error) {
    if len(args) < 1 {
        return "", fmt.Errorf("jiraSearchTasks command requires at least 1 argument")
    }
    jiraTasks, err := c.makeJiraSearch(args[0])
    if err != nil {
        return "", fmt.Errorf("failed to make jira search: %w", err)
    }

    description, err := formatTasks(jiraTasks)
    if err != nil {
        return "", fmt.Errorf("failed to describe json with chat gpt: %w", err)
    }

    return description, nil
}

func (c *CommandJiraGetTasks) makeJiraSearch(jql string) (*jiraSearchResponseSpec, error) {
    uri := fmt.Sprintf(
        "https://%s/rest/api/3/search?jql=%s&fields=summary,priority,assignee,status,creator,subtasks,issuetype,project,description,comment,parent&maxResults=10",
        c.host,
        url.QueryEscape(jql),
    )
    log.Debugf("jira url: %s", uri)
    client := &http.Client{}
    req, err := http.NewRequest("GET", uri, nil)
    reqToken := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.user, c.token)))
    req.Header.Set("Authorization", fmt.Sprintf("Basic %s", reqToken))
    req.Header.Set("Content-Type", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to create jira client: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read jira response: %w", err)
    }

    var spec jiraSearchResponseSpec
    if err := json.Unmarshal(body, &spec); err != nil {
        return nil, fmt.Errorf("failed to unmarshal jira response: %v", err)
    }

    return &spec, nil
}

func formatTasks(js *jiraSearchResponseSpec) (string, error) {
    spew.Dump(js)
    t, err := template.New("person").Parse(`
{{range .Issues}}
    - {{.Key}} {{.Fields.Issuetype.Name}} {{.Fields.Summary}}({{.Fields.Creator.DisplayName}} -> {{.Fields.Assignee.DisplayName}}]): {{.Fields.Status.Name}} 
{{- end}}
`)
    if err != nil {
        return "", fmt.Errorf("failed to parse template: %w", err)
    }

    var buf bytes.Buffer
    err = t.Execute(&buf, js)
    if err != nil {
        return "", fmt.Errorf("failed to execute template: %w", err)
    }

    return buf.String(), nil
}

type jiraSearchResponseSpec struct {
    StartAt    int `json:"startAt"`
    MaxResults int `json:"maxResults"`
    Total      int `json:"total"`
    Issues     []struct {
        ID     string `json:"id"`
        Key    string `json:"key"`
        Fields struct {
            Summary   string `json:"summary"`
            Issuetype struct {
                Description string `json:"description"`
                //IconURL     string `json:"iconUrl"`
                Name string `json:"name"`
            } `json:"issuetype"`
            Creator struct {
                EmailAddress string `json:"emailAddress"`
                //AvatarUrls   struct {
                //    Four8X48 string `json:"48x48"`
                //} `json:"avatarUrls"`
                DisplayName string `json:"displayName"`
            } `json:"creator"`
            Project struct {
                Key  string `json:"key"`
                Name string `json:"name"`
                //AvatarUrls struct {
                //    Four8X48 string `json:"48x48"`
                //} `json:"avatarUrls"`
            } `json:"project"`
            Assignee struct {
                EmailAddress string `json:"emailAddress"`
                //AvatarUrls   struct {
                //    Four8X48 string `json:"48x48"`
                //} `json:"avatarUrls"`
                DisplayName string `json:"displayName"`
            } `json:"assignee"`
            Priority struct {
                //IconURL string `json:"iconUrl"`
                Name string `json:"name"`
            } `json:"priority"`
            Status struct {
                Name           string `json:"name"`
                ID             string `json:"id"`
                StatusCategory struct {
                    Key  string `json:"key"`
                    Name string `json:"name"`
                } `json:"statusCategory"`
            } `json:"status"`
        } `json:"fields"`
    } `json:"issues"`
}
