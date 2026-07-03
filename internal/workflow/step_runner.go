package workflow

type StepRunner struct {
	Step    string
	Command string
}

func (r StepRunner) Ready() bool {
	return r.Step != "" && r.Command != ""
}
