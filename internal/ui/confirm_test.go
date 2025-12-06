package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfirmer_Confirm_YesLowercase(t *testing.T) {
	input := strings.NewReader("y\n")
	output := &bytes.Buffer{}
	confirmer := &Confirmer{In: input, Out: output}

	result := confirmer.Confirm("Delete? [y/N]: ")

	assert.Equal(t, ConfirmYes, result)
}

func TestConfirmer_Confirm_YesUppercase(t *testing.T) {
	input := strings.NewReader("Y\n")
	output := &bytes.Buffer{}
	confirmer := &Confirmer{In: input, Out: output}

	result := confirmer.Confirm("Delete? [y/N]: ")

	assert.Equal(t, ConfirmYes, result)
}

func TestConfirmer_Confirm_YesFull(t *testing.T) {
	input := strings.NewReader("yes\n")
	output := &bytes.Buffer{}
	confirmer := &Confirmer{In: input, Out: output}

	result := confirmer.Confirm("Delete? [y/N]: ")

	assert.Equal(t, ConfirmYes, result)
}

func TestConfirmer_Confirm_YesFullMixedCase(t *testing.T) {
	input := strings.NewReader("YeS\n")
	output := &bytes.Buffer{}
	confirmer := &Confirmer{In: input, Out: output}

	result := confirmer.Confirm("Delete? [y/N]: ")

	assert.Equal(t, ConfirmYes, result)
}

func TestConfirmer_Confirm_NoLowercase(t *testing.T) {
	input := strings.NewReader("n\n")
	output := &bytes.Buffer{}
	confirmer := &Confirmer{In: input, Out: output}

	result := confirmer.Confirm("Delete? [y/N]: ")

	assert.Equal(t, ConfirmNo, result)
}

func TestConfirmer_Confirm_EmptyInputDefaultsNo(t *testing.T) {
	input := strings.NewReader("\n")
	output := &bytes.Buffer{}
	confirmer := &Confirmer{In: input, Out: output}

	result := confirmer.Confirm("Delete? [y/N]: ")

	assert.Equal(t, ConfirmNo, result, "empty input should default to No")
}

func TestConfirmer_Confirm_InvalidInputDefaultsNo(t *testing.T) {
	input := strings.NewReader("maybe\n")
	output := &bytes.Buffer{}
	confirmer := &Confirmer{In: input, Out: output}

	result := confirmer.Confirm("Delete? [y/N]: ")

	assert.Equal(t, ConfirmNo, result, "invalid input should default to No")
}

func TestConfirmer_Confirm_WhitespaceOnlyDefaultsNo(t *testing.T) {
	input := strings.NewReader("   \n")
	output := &bytes.Buffer{}
	confirmer := &Confirmer{In: input, Out: output}

	result := confirmer.Confirm("Delete? [y/N]: ")

	assert.Equal(t, ConfirmNo, result, "whitespace input should default to No")
}

func TestConfirmer_Confirm_DisplaysPrompt(t *testing.T) {
	input := strings.NewReader("n\n")
	output := &bytes.Buffer{}
	confirmer := &Confirmer{In: input, Out: output}

	confirmer.Confirm("Delete? [y/N]: ")

	assert.Contains(t, output.String(), "Delete? [y/N]:")
}

func TestConfirmer_Confirm_EOFReturnsNo(t *testing.T) {
	input := strings.NewReader("") // EOF immediately
	output := &bytes.Buffer{}
	confirmer := &Confirmer{In: input, Out: output}

	result := confirmer.Confirm("Delete? [y/N]: ")

	assert.Equal(t, ConfirmNo, result, "EOF should return No")
}

func TestConfirmChanges_DisplaysPreview(t *testing.T) {
	preview := &Preview{
		Title: "Test Preview",
		Changes: []Change{
			{Action: ActionDelete, Path: "/path/to/delete", Size: 1000},
		},
	}

	input := strings.NewReader("n\n")
	output := &bytes.Buffer{}

	_, err := ConfirmChanges(preview, input, output, false)
	require.NoError(t, err)

	assert.Contains(t, output.String(), "Test Preview")
}

func TestConfirmChanges_ReturnsTrueOnYes(t *testing.T) {
	preview := &Preview{
		Title:   "Test",
		Changes: []Change{{Action: ActionDelete, Path: "/test", Size: 100}},
	}

	input := strings.NewReader("y\n")
	output := &bytes.Buffer{}

	confirmed, err := ConfirmChanges(preview, input, output, false)
	require.NoError(t, err)

	assert.True(t, confirmed)
}

func TestConfirmChanges_ReturnsFalseOnNo(t *testing.T) {
	preview := &Preview{
		Title:   "Test",
		Changes: []Change{{Action: ActionDelete, Path: "/test", Size: 100}},
	}

	input := strings.NewReader("n\n")
	output := &bytes.Buffer{}

	confirmed, err := ConfirmChanges(preview, input, output, false)
	require.NoError(t, err)

	assert.False(t, confirmed)
}

func TestConfirmChanges_AutoYesSkipsPrompt(t *testing.T) {
	preview := &Preview{
		Title:   "Test",
		Changes: []Change{{Action: ActionDelete, Path: "/test", Size: 100}},
	}

	input := strings.NewReader("") // No input provided
	output := &bytes.Buffer{}

	confirmed, err := ConfirmChanges(preview, input, output, true)
	require.NoError(t, err)

	assert.True(t, confirmed, "autoYes should return true without prompting")
	assert.Contains(t, output.String(), "Test", "should still display preview with autoYes")
}

func TestConfirmChanges_ShowsAbortedMessage(t *testing.T) {
	preview := &Preview{
		Title:   "Test",
		Changes: []Change{{Action: ActionDelete, Path: "/test", Size: 100}},
	}

	input := strings.NewReader("n\n")
	output := &bytes.Buffer{}

	_, _ = ConfirmChanges(preview, input, output, false)

	assert.Contains(t, output.String(), "Aborted")
}
