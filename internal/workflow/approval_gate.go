package workflow

import "fmt"

func RequireApproval(step string, approved bool) error {
	if approved {
		return nil
	}
	return fmt.Errorf("%s requires user approval before next step", step)
}
