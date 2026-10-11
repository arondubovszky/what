package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/arondubovszky/what/form"
)

type rawQuestion struct {
	Type    string   `json:"type"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options"`
	Correct []int    `json:"correct"`
}

// raw JSON so property order is preserved: Go maps marshal keys alphabetically,
// which made the model emit "correct" before writing the question
var questionsSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"questions": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"type": {
						"type": "string",
						"enum": ["select", "multiple", "binary"],
						"description": "select: exactly one correct answer among 3-5 options. multiple: 3-5 options, one or more correct. binary: a true/false statement."
					},
					"prompt": {
						"type": "string",
						"description": "The question text only. Never list the answer choices here; they go in options."
					},
					"options": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Answer choices, one per string, without letter/number prefixes. select and multiple: 3-5 entries, never empty. binary: exactly [\"true\", \"false\"]."
					},
					"correct": {
						"type": "array",
						"items": {"type": "integer"},
						"description": "0-based indices into options. select and binary: exactly one index. multiple: one or more indices. Every index must be less than len(options)."
					}
				},
				"required": ["type", "prompt", "options", "correct"],
				"additionalProperties": false
			}
		}
	},
	"required": ["questions"],
	"additionalProperties": false
}`)

func Generate(ctx context.Context, client anthropic.Client, material string, n int) ([]form.Question, error) {
	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-opus-5-5",
		MaxTokens: 4096,
		System:    []anthropic.TextBlockParam{{Text: "You write quiz questions. Pick whichever question type fits each concept best."}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				fmt.Sprintf("Write exactly %d questions, each with a non-empty prompt, about:\n\n%s", n, material))),
		},
	}, option.WithJSONSet("output_config", map[string]any{
		"format": map[string]any{"type": "json_schema", "schema": questionsSchema},
	}))
	if err != nil {
		return nil, err
	}
	if msg.StopReason != "end_turn" {
		return nil, fmt.Errorf("unexpected stop_reason %s", msg.StopReason) // max_tokens = truncated JSON
	}

	for _, block := range msg.Content {
		if block.Type != "text" {
			continue
		}
		var out struct{ Questions []rawQuestion }
		if err := json.Unmarshal([]byte(block.Text), &out); err != nil {
			return nil, err
		}
		return convert(out.Questions)
	}
	return nil, fmt.Errorf("no text block in response")
}

func convert(raw []rawQuestion) ([]form.Question, error) {
	types := map[string]form.QuestionType{"select": form.Select, "multiple": form.Multiple, "binary": form.Binary}
	var qs []form.Question
	for i, r := range raw {
		t, ok := types[r.Type]
		if !ok {
			return nil, fmt.Errorf("q%d: unknown type %q", i, r.Type)
		}
		if t == form.Binary {
			r.Options = []string{"true", "false"}
		}
		if len(r.Correct) == 0 || (t != form.Multiple && len(r.Correct) != 1) {
			return nil, fmt.Errorf("q%d: bad correct count %d for %s", i, len(r.Correct), r.Type)
		}
		for _, c := range r.Correct {
			if c < 0 || c >= len(r.Options) {
				return nil, fmt.Errorf("q%d: correct index %d out of range", i, c)
			}
		}
		qs = append(qs, form.Question{Type: t, Prompt: r.Prompt, Options: r.Options, Correct: r.Correct})
	}
	return qs, nil
}
