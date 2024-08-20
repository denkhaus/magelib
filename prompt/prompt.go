package prompt

import (
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

// Input displays a prompt with a given label and a list of choices, and returns the selected choice.
//
// Parameters:
// - label: the label to display for the prompt.
// - validate: a function that validates the user's input.
//
// Returns:
// - result: the user's input.
// - error: an error if the prompt failed.
func Input(label string, validate func(input string) error) (string, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", errors.Wrap(err, "prompt failed")
	}

	return result, nil
}

// YesNo prompts the user with a yes/no question and returns the user's response as a boolean value.
//
// The parameter `label` is a string that represents the question to be prompted to the user.
// The function returns a boolean value indicating whether the user answered "yes" or not, and an error if the prompt fails.
func YesNo(label string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		return false, errors.Wrap(err, "prompt failed")
	}

	return result == "yes", nil
}

// Select displays a prompt with a given label and a list of choices, and returns the selected choice.
//
// Parameters:
// - label: the label to display for the prompt.
// - choices: the list of choices to display in the prompt.
//
// Returns:
// - result: the selected choice.
// - error: an error if the prompt failed.
func Select(label string, choices []string) (string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: choices,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", errors.Wrap(err, "prompt failed")
	}

	return result, nil
}
