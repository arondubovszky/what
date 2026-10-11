package form

import (
	"fmt"
	"slices"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

var (
	prompStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	answerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#02BF87")).Bold(true)
)

func buildOptions(opts []string) []huh.Option[int] {
	options := make([]huh.Option[int], len(opts))
	for i, opt := range opts {
		options[i] = huh.NewOption(opt, i)
	}

	return options
}

func buildMultipleChoice(q Question, answers *[]int) *huh.MultiSelect[int] {
	options := buildOptions(q.Options)

	return huh.NewMultiSelect[int]().Title(q.Prompt).Options(options...).Value(answers)
}

func buildSelect(q Question, answer *int) *huh.Select[int] {
	options := buildOptions(q.Options)

	return huh.NewSelect[int]().Title(q.Prompt).Options(options...).Value(answer)
}

func buildBinary(q Question, answer *int) *huh.Select[int] {
	return huh.NewSelect[int]().Title(q.Prompt).Options(buildOptions(q.Options)...).Value(answer)
}

func createGroup(questions []Question, answers [][]int) *huh.Group {
	var fields = make([]huh.Field, len(questions))
	for i, q := range questions {
		if q.Type != Multiple {
			answers[i] = make([]int, 1)
		}
		switch q.Type {
		case Select:
			fields[i] = buildSelect(q, &answers[i][0])
		case Multiple:
			fields[i] = buildMultipleChoice(q, &answers[i])
		case Binary:
			fields[i] = buildBinary(q, &answers[i][0])
		}
	}

	return huh.NewGroup(fields...)
}

func buildMistakeNote(m Mistake, idx int) huh.Field {
	desc := fmt.Sprintf("%s %s\n%s %s\n%s %s",
		prompStyle.Render("question:"), m.Prompt,
		answerStyle.Render("got:"), m.Answer,
		correctStyle.Render("correct:"), m.Correct,
	)
	return huh.NewNote().Title(fmt.Sprintf("Question %d", idx)).Description(desc)
}

// evaluates aswell
func Form(questions []Question, groupLen int) error {
	answers := make([][]int, len(questions))
	var groups []*huh.Group
	for start := 0; start < len(questions); start += groupLen {
		end := min(start+groupLen, len(questions))
		groups = append(groups, createGroup(questions[start:end], answers[start:end]))
	}

	if err := huh.NewForm(groups...).Run(); err != nil {
		return err
	}

	// printed to stdout instead of a huh group so long summaries stay scrollable
	mistakes := 0
	for i, q := range questions {
		if !sameSlices(answers[i], q.Correct) {
			mistakes++
			lipgloss.Println(renderMistake(Mistake{Prompt: q.Prompt, Answer: q.concatOptions(answers[i]), Correct: q.concatOptions(q.Correct)}, i+1))
		}
	}

	title := fmt.Sprintf("Summary - %d/%d", len(questions)-mistakes, len(questions))
	if mistakes == 0 {
		title = "Summary - 100%"
	}

	var done bool
	confirm := huh.NewConfirm().Title(title).Affirmative("Done").Negative("Restart").Value(&done)
	if err := huh.NewForm(huh.NewGroup(confirm)).Run(); err != nil {
		return err
	}

	if !done {
		if err := Form(questions, groupLen); err != nil {
			return err
		}
	}

	return nil
}

func sameSlices(a, b []int) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

func renderMistake(m Mistake, idx int) string {
	return fmt.Sprintf("%s\n%s %s\n%s %s\n%s %s\n",
		lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Question %d", idx)),
		prompStyle.Render("question:"), m.Prompt,
		answerStyle.Render("got:"), m.Answer,
		correctStyle.Render("correct:"), m.Correct,
	)
}
