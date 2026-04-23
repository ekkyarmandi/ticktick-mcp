package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const defaultTickTickAPIBase = "https://api.ticktick.com/open/v1"

type tickTickClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func newTickTickClientFromEnv() (*tickTickClient, error) {
	apiKey := strings.TrimSpace(os.Getenv("TICKTICK_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("TICKTICK_API_KEY is required")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("TICKTICK_API_BASE")), "/")
	if baseURL == "" {
		baseURL = defaultTickTickAPIBase
	}

	return &tickTickClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *tickTickClient) request(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		bodyText := strings.TrimSpace(string(data))
		if bodyText == "" {
			bodyText = http.StatusText(resp.StatusCode)
		}
		return nil, fmt.Errorf("ticktick %s %s failed with %d: %s", method, path, resp.StatusCode, bodyText)
	}

	return data, nil
}

func decodeJSONObject(data []byte) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func decodeJSONArray(data []byte) ([]map[string]any, error) {
	var out []map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func parseTickTickDate(raw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.999-0700",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported TickTick date format: %q", raw)
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

type tickTickService struct {
	client             *tickTickClient
	excludedGroupIDs   map[string]bool
	excludedProjectIDs map[string]bool
}

type noArgs struct{}

type projectIDInput struct {
	ProjectID string `json:"project_id" jsonschema:"TickTick project ID"`
}

type projectTaskInput struct {
	ProjectID string `json:"project_id" jsonschema:"TickTick project ID"`
	TaskID    string `json:"task_id" jsonschema:"TickTick task ID"`
}

type todayTasksInput struct {
	Timezone string `json:"timezone,omitempty" jsonschema:"IANA timezone name used to calculate today; defaults to Asia/Jakarta"`
}

type createProjectInput struct {
	ProjectName string `json:"project_name" jsonschema:"Name of the project to create"`
}

type projectsOutput struct {
	Projects []map[string]any `json:"projects" jsonschema:"TickTick projects"`
}

type todayTasksOutput struct {
	Tasks []map[string]any `json:"tasks" jsonschema:"TickTick tasks due today"`
}

type createTaskInput struct {
	ProjectID  string           `json:"project_id" jsonschema:"TickTick project ID"`
	Title      string           `json:"title" jsonschema:"Task title"`
	Content    *string          `json:"content,omitempty" jsonschema:"Task content"`
	Desc       *string          `json:"desc,omitempty" jsonschema:"Checklist description"`
	IsAllDay   *bool            `json:"isAllDay,omitempty" jsonschema:"Whether the task is an all-day event"`
	StartDate  *string          `json:"startDate,omitempty" jsonschema:"Start date/time in TickTick format, for example 2019-11-13T03:00:00+0000"`
	DueDate    *string          `json:"dueDate,omitempty" jsonschema:"Due date/time in TickTick format, for example 2019-11-13T03:00:00+0000"`
	TimeZone   *string          `json:"timeZone,omitempty" jsonschema:"IANA time zone associated with the task dates"`
	Reminders  []string         `json:"reminders,omitempty" jsonschema:"Reminder trigger strings"`
	RepeatFlag *string          `json:"repeatFlag,omitempty" jsonschema:"TickTick repeat rule"`
	Priority   int              `json:"priority" jsonschema:"Task priority, where 0 is normal"`
	SortOrder  *int             `json:"sortOrder,omitempty" jsonschema:"Sort order within the project"`
	Items      []map[string]any `json:"items,omitempty" jsonschema:"Subtasks or checklist items"`
}

type updateTaskInput struct {
	TaskID     string           `json:"task_id" jsonschema:"TickTick task ID"`
	ProjectID  string           `json:"project_id" jsonschema:"TickTick project ID"`
	Title      *string          `json:"title,omitempty" jsonschema:"Task title"`
	Content    *string          `json:"content,omitempty" jsonschema:"Task content"`
	Desc       *string          `json:"desc,omitempty" jsonschema:"Checklist description"`
	IsAllDay   *bool            `json:"isAllDay,omitempty" jsonschema:"Whether the task is an all-day event"`
	StartDate  *string          `json:"startDate,omitempty" jsonschema:"Start date/time in TickTick format, for example 2019-11-13T03:00:00+0000"`
	DueDate    *string          `json:"dueDate,omitempty" jsonschema:"Due date/time in TickTick format, for example 2019-11-13T03:00:00+0000"`
	TimeZone   *string          `json:"timeZone,omitempty" jsonschema:"IANA time zone associated with the task dates"`
	Reminders  []string         `json:"reminders,omitempty" jsonschema:"Reminder trigger strings"`
	RepeatFlag *string          `json:"repeatFlag,omitempty" jsonschema:"TickTick repeat rule"`
	Priority   *int             `json:"priority,omitempty" jsonschema:"Task priority, where 0 is normal"`
	SortOrder  *int             `json:"sortOrder,omitempty" jsonschema:"Sort order within the project"`
	Items      []map[string]any `json:"items,omitempty" jsonschema:"Subtasks or checklist items"`
}

func registerTools(server *mcp.Server, service *tickTickService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_projects",
		Description: "Get a list of TickTick projects.",
	}, service.getProjects)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_details",
		Description: "Get detailed project data, including tasks, for a TickTick project.",
	}, service.projectDetails)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_today_tasks",
		Description: "Get all TickTick tasks due today across projects.",
	}, service.getTodayTasks)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task_details",
		Description: "Get a single TickTick task by project and task ID.",
	}, service.getTaskDetails)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_project",
		Description: "Create a new TickTick project.",
	}, service.createProject)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_task",
		Description: "Create a new task in TickTick.",
	}, service.createTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_task",
		Description: "Update an existing TickTick task.",
	}, service.updateTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "complete_task",
		Description: "Mark a TickTick task as complete.",
	}, service.completeTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_task",
		Description: "Delete a TickTick task.",
	}, service.deleteTask)
}

func (s *tickTickService) getProjects(ctx context.Context, _ *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, projectsOutput, error) {
	data, err := s.client.request(ctx, http.MethodGet, "/project", nil)
	if err != nil {
		return nil, projectsOutput{}, err
	}

	projects, err := decodeJSONArray(data)
	if err == nil {
		return nil, projectsOutput{Projects: projects}, nil
	}

	var wrapped struct {
		Projects []map[string]any `json:"projects"`
	}
	if unwrapErr := json.Unmarshal(data, &wrapped); unwrapErr == nil && wrapped.Projects != nil {
		return nil, projectsOutput{Projects: wrapped.Projects}, nil
	}

	return nil, projectsOutput{}, fmt.Errorf("unexpected project list response")
}

func (s *tickTickService) projectDetails(ctx context.Context, _ *mcp.CallToolRequest, input projectIDInput) (*mcp.CallToolResult, map[string]any, error) {
	data, err := s.client.request(ctx, http.MethodGet, "/project/"+input.ProjectID+"/data", nil)
	if err != nil {
		return nil, nil, err
	}

	out, err := decodeJSONObject(data)
	if err != nil {
		return nil, nil, err
	}

	return nil, out, nil
}

func (s *tickTickService) getTodayTasks(ctx context.Context, _ *mcp.CallToolRequest, input todayTasksInput) (*mcp.CallToolResult, todayTasksOutput, error) {
	timezone := input.Timezone
	if strings.TrimSpace(timezone) == "" {
		timezone = "Asia/Jakarta"
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, todayTasksOutput{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}

	now := time.Now().In(location)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	todayEnd := todayStart.Add(24 * time.Hour)

	projects, err := s.getProjectsList(ctx)
	if err != nil {
		return nil, todayTasksOutput{}, err
	}

	todayTasks := make([]map[string]any, 0)
	for _, project := range projects {
		projectID := stringValue(project["id"])
		if projectID == "" {
			continue
		}

		groupID := stringValue(project["groupId"])
		if s.excludedGroupIDs[groupID] || s.excludedProjectIDs[projectID] {
			continue
		}

		projectName := stringValue(project["name"])
		projectData, err := s.getProjectData(ctx, projectID)
		if err != nil {
			return nil, todayTasksOutput{}, err
		}

		for _, task := range projectData.Tasks {
			rawDate := stringValue(task["dueDate"])
			if rawDate == "" {
				rawDate = stringValue(task["startDate"])
			}
			if rawDate == "" {
				continue
			}

			parsed, err := parseTickTickDate(rawDate)
			if err != nil {
				continue
			}

			localTime := parsed.In(location)
			if localTime.Before(todayStart) || !localTime.Before(todayEnd) {
				continue
			}

			task["_projectName"] = projectName
			todayTasks = append(todayTasks, task)
		}
	}

	return nil, todayTasksOutput{Tasks: todayTasks}, nil
}

func (s *tickTickService) getTaskDetails(ctx context.Context, _ *mcp.CallToolRequest, input projectTaskInput) (*mcp.CallToolResult, map[string]any, error) {
	data, err := s.client.request(ctx, http.MethodGet, "/project/"+input.ProjectID+"/task/"+input.TaskID, nil)
	if err != nil {
		return nil, nil, err
	}

	out, err := decodeJSONObject(data)
	if err != nil {
		return nil, nil, err
	}

	return nil, out, nil
}

func (s *tickTickService) createProject(ctx context.Context, _ *mcp.CallToolRequest, input createProjectInput) (*mcp.CallToolResult, map[string]any, error) {
	payload := map[string]any{"name": input.ProjectName}

	data, err := s.client.request(ctx, http.MethodPost, "/project", payload)
	if err != nil {
		return nil, nil, err
	}

	out, err := decodeJSONObject(data)
	if err != nil {
		return nil, nil, err
	}

	return nil, out, nil
}

func (s *tickTickService) createTask(ctx context.Context, _ *mcp.CallToolRequest, input createTaskInput) (*mcp.CallToolResult, map[string]any, error) {
	payload := map[string]any{
		"projectId": input.ProjectID,
		"title":     input.Title,
		"priority":  input.Priority,
	}
	addTaskPayloadFields(payload, input.Content, input.Desc, input.IsAllDay, input.StartDate, input.DueDate, input.TimeZone, input.Reminders, input.RepeatFlag, input.SortOrder, input.Items)

	data, err := s.client.request(ctx, http.MethodPost, "/task", payload)
	if err != nil {
		return nil, nil, err
	}

	out, err := decodeJSONObject(data)
	if err != nil {
		return nil, nil, err
	}

	return nil, out, nil
}

func (s *tickTickService) updateTask(ctx context.Context, _ *mcp.CallToolRequest, input updateTaskInput) (*mcp.CallToolResult, map[string]any, error) {
	payload := map[string]any{
		"id":        input.TaskID,
		"projectId": input.ProjectID,
	}

	if input.Title != nil {
		payload["title"] = *input.Title
	}
	if input.Priority != nil {
		payload["priority"] = *input.Priority
	}
	addTaskPayloadFields(payload, input.Content, input.Desc, input.IsAllDay, input.StartDate, input.DueDate, input.TimeZone, input.Reminders, input.RepeatFlag, input.SortOrder, input.Items)

	data, err := s.client.request(ctx, http.MethodPost, "/task/"+input.TaskID, payload)
	if err != nil {
		return nil, nil, err
	}

	out, err := decodeJSONObject(data)
	if err != nil {
		return nil, nil, err
	}

	return nil, out, nil
}

func (s *tickTickService) completeTask(ctx context.Context, _ *mcp.CallToolRequest, input projectTaskInput) (*mcp.CallToolResult, map[string]string, error) {
	if _, err := s.client.request(ctx, http.MethodPost, "/project/"+input.ProjectID+"/task/"+input.TaskID+"/complete", nil); err != nil {
		return nil, nil, err
	}

	return nil, map[string]string{"message": "Task completed"}, nil
}

func (s *tickTickService) deleteTask(ctx context.Context, _ *mcp.CallToolRequest, input projectTaskInput) (*mcp.CallToolResult, map[string]string, error) {
	if _, err := s.client.request(ctx, http.MethodDelete, "/project/"+input.ProjectID+"/task/"+input.TaskID, nil); err != nil {
		return nil, nil, err
	}

	return nil, map[string]string{"message": "Task deleted"}, nil
}

func (s *tickTickService) getProjectsList(ctx context.Context) ([]map[string]any, error) {
	data, err := s.client.request(ctx, http.MethodGet, "/project", nil)
	if err != nil {
		return nil, err
	}

	projects, err := decodeJSONArray(data)
	if err == nil {
		return projects, nil
	}

	var wrapped struct {
		Projects []map[string]any `json:"projects"`
	}
	if unwrapErr := json.Unmarshal(data, &wrapped); unwrapErr == nil && wrapped.Projects != nil {
		return wrapped.Projects, nil
	}

	return nil, fmt.Errorf("unexpected project list response")
}

type projectData struct {
	Tasks []map[string]any `json:"tasks"`
}

func (s *tickTickService) getProjectData(ctx context.Context, projectID string) (*projectData, error) {
	data, err := s.client.request(ctx, http.MethodGet, "/project/"+projectID+"/data", nil)
	if err != nil {
		return nil, err
	}

	var out projectData
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func addTaskPayloadFields(
	payload map[string]any,
	content *string,
	desc *string,
	isAllDay *bool,
	startDate *string,
	dueDate *string,
	timeZone *string,
	reminders []string,
	repeatFlag *string,
	sortOrder *int,
	items []map[string]any,
) {
	if content != nil {
		payload["content"] = *content
	}
	if desc != nil {
		payload["desc"] = *desc
	}
	if isAllDay != nil {
		payload["isAllDay"] = *isAllDay
	}
	if startDate != nil {
		payload["startDate"] = *startDate
	}
	if dueDate != nil {
		payload["dueDate"] = *dueDate
	}
	if timeZone != nil {
		payload["timeZone"] = *timeZone
	}
	if len(reminders) > 0 {
		payload["reminders"] = reminders
	}
	if repeatFlag != nil {
		payload["repeatFlag"] = *repeatFlag
	}
	if sortOrder != nil {
		payload["sortOrder"] = *sortOrder
	}
	if len(items) > 0 {
		payload["items"] = items
	}
}
