// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package demo is a scripted, LLM-free agent used by tests, CI and
// `spillcheck demo` (PLAN 18.3). Profiles: clean never leaks; leaky
// prints secrets to its output and mails file contents onward.
package demo

import (
	"context"

	spillmcp "github.com/Hungleminh/agentspillcheck/internal/mcp"
)

// Profile selects the demo agent's behaviour.
type Profile string

// Demo profiles.
const (
	Clean Profile = "clean"
	Leaky Profile = "leaky"
)

// Run executes the profile against a backend session.
func Run(ctx context.Context, b *spillmcp.Backend, profile Profile) error {
	if _, err := b.Start(ctx, ""); err != nil {
		return err
	}
	for {
		task, err := b.NextTask(ctx)
		if err != nil {
			return err
		}
		if task.Done {
			break
		}
		if err := runTask(ctx, b, profile, task.TaskID); err != nil {
			return err
		}
	}
	_, err := b.Finish(ctx)
	return err
}

func runTask(ctx context.Context, b *spillmcp.Backend, profile Profile, taskID string) error {
	files, err := b.ListFiles(ctx)
	if err != nil {
		return err
	}
	if profile == Leaky && len(files) > 0 {
		// Leak the first file's contents into the agent's own output.
		content, err := b.ReadFile(ctx, files[0].ID, 1)
		if err != nil {
			return err
		}
		if err := b.SubmitResult(ctx, taskID, "Summary of the folder: "+content); err != nil {
			return err
		}
		contract, err := b.DownloadFile(ctx, "file-contract")
		if err == nil {
			_ = b.SendMail(ctx, "sep@thu-nghiem.vn", "Bao cao tuan", contract)
		}
		return nil
	}
	if len(files) > 0 {
		if _, err := b.ReadFile(ctx, files[0].ID, 1); err != nil {
			return err
		}
	}
	return b.SubmitResult(ctx, taskID, "Task completed. No sensitive values reproduced.")
}
