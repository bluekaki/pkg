package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/mm/httpclient"
)

type _Label struct {
	Job          string `json:"job"`
	Project      string `json:"project"`
	PodName      string `json:"podname"`
	JobCreatedAt *_Time `json:"job_created_at"`
}

func (l *_Label) String() string {
	return fmt.Sprintf("job: %s, project: %s, podname: %s, job_created_at: %s", l.Job, l.Project, l.PodName, l.JobCreatedAt)
}

type _Time struct {
	val time.Time
}

func (t *_Time) MarshalJSON() ([]byte, error) {
	return []byte(t.val.Format(time.RFC3339)), nil
}

func (t *_Time) UnmarshalJSON(raw []byte) error {
	ts, err := time.Parse(time.RFC3339, string(raw[1:len(raw)-1]))
	if err != nil {
		return err
	}

	t.val = ts
	return nil
}

func (t *_Time) String() string {
	val, _ := t.MarshalJSON()
	return string(val)
}

// http://127.0.0.1:9091/api/v1/metrics
func ParseMetrics(metrics []byte) ([]byte, error) {
	targets := &struct {
		Data []struct {
			Label *_Label `json:"labels"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(metrics, targets); err != nil {
		return nil, errors.Wrap(err, "unmarshal json format metrics err")
	}

	buffer := bytes.NewBuffer(nil)
	groups := make(map[string][]*_Label)

	for _, target := range targets.Data {
		if target.Label.Job != "metrics-porter" {
			buffer.WriteString(target.Label.String())
			buffer.WriteString("\n")
			continue
		}

		groups[target.Label.Project] = append(groups[target.Label.Project], target.Label)
	}

	if buffer.Len() > 0 {
		buffer.WriteString("\n")
		buffer.WriteString("--------------------------------------------------")
		buffer.WriteString("\n\n")
	}

	projects := make([]string, 0, len(groups))
	for project := range groups {
		projects = append(projects, project)
	}

	sort.Strings(projects)
	for _, project := range projects {
		labels := groups[project]
		sort.Slice(labels, func(i, j int) bool {
			return labels[i].JobCreatedAt.val.Before(labels[j].JobCreatedAt.val)
		})

		for i, labels := range labels {
			buffer.WriteString(fmt.Sprintf("[%d]  %s", i, labels.String()))
			buffer.WriteString("\n")
		}
		buffer.WriteString(fmt.Sprintf("{%.2f minutes}\n", labels[len(labels)-1].JobCreatedAt.val.Sub(labels[0].JobCreatedAt.val).Minutes()))

		buffer.WriteString("\n")
		buffer.WriteString("--------------------------------------------------")
		buffer.WriteString("\n\n")
	}

	return buffer.Bytes(), nil
}

// 127.0.0.1:9091
func DeleteAll(metricServer string) error {
	body, _, _, err := httpclient.Get(fmt.Sprintf("http://%s/api/v1/metrics", metricServer), nil,
		httpclient.WithRetryTimes(0), httpclient.WithTTL(time.Minute))
	if err != nil {
		return errors.Wrap(err, "get metrics err")
	}

	targets := &struct {
		Data []struct {
			Labels map[string]string `json:"labels"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(body, targets); err != nil {
		return errors.Wrap(err, "unmarshal json format metrics err")
	}

	for _, target := range targets.Data {
		job, ok := target.Labels["job"]
		if !ok {
			continue
		}
		delete(target.Labels, "job")

		lables := make([]string, 0, len(target.Labels))
		for name, value := range target.Labels {
			lables = append(lables, fmt.Sprintf("%s/%s", name, value))
		}

		sort.Strings(lables)
		url := fmt.Sprintf("http://%s/metrics/job/%s/%s", metricServer, job, strings.Join(lables, "/"))

		_, _, statusCode, err := httpclient.Delete(url, nil, httpclient.WithRetryTimes(0), httpclient.WithTTL(time.Minute))
		if statusCode != http.StatusAccepted && err != nil {
			return errors.Wrap(err, "delete metrics err")
		}
	}

	return nil
}
