package form

import "strings"

type QuestionType int

const (
	Select   QuestionType = iota // ABCD -> NewGroup
	Multiple                     // multiple choice -> MultiSelect
	Binary                       // true/fal
)

type Question struct {
	Type    QuestionType
	Prompt  string
	Options []string // Binary: ["true", "false"]
	Correct []int
}

func (q Question) concatOptions(idx []int) string {
	res := make([]string, len(idx))
	for i, j := range idx {
		res[i] = q.Options[j]
	}
	return strings.Join(res, ", ")
}

type Mistake struct {
	Prompt, Answer, Correct string
}
