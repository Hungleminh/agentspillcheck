package completion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
	"github.com/Hungleminh/agentspillcheck/internal/report"
)

// Reporter delivers only the redacted server-generated report to Marketplace.
type Reporter struct {
	URL, Token string
	Client     *http.Client
}

func (r Reporter) Complete(ctx context.Context, marketplaceSessionID string, backend *spillmcp.Backend) error {
	if r.URL == "" || r.Token == "" {
		return nil
	}
	rep := report.Build(backend.SessionID(), backend.Baits(), backend.Events())
	raw, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	html, err := report.RenderHTML(rep, "en")
	if err != nil {
		return err
	}
	markdown, err := report.RenderMarkdown(rep, "en")
	if err != nil {
		return err
	}
	tasksComplete := len(rep.Tasks) == 8
	for _, task := range rep.Tasks {
		tasksComplete = tasksComplete && task.Status == "completed"
	}
	payload := map[string]any{"score": rep.Score, "ranking_eligible": tasksComplete, "report_json": string(raw), "report_html": html, "report_markdown": markdown, "findings": findings(rep)}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(r.URL, "/")+"/api/internal/spill-check/sessions/"+marketplaceSessionID+"/complete", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.Token)
	req.Header.Set("Content-Type", "application/json")
	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("marketplace completion callback: %s", resp.Status)
	}
	return nil
}

func findings(rep report.Report) []map[string]string {
	out := make([]map[string]string, 0, len(rep.Findings))
	for _, finding := range rep.Findings {
		out = append(out, map[string]string{"rule_id": finding.Subclass, "scenario_id": finding.Carrier, "severity": finding.Severity})
	}
	return out
}
