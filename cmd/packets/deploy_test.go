package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeployVerdict_anExplicitHostCommandWithNoCheckIsItsOwnEvidence(t *testing.T) {
	t.Parallel()

	status, refusal := deployVerdict("deployed", false, nil)
	assert.Equal(t, "deployed", status)
	assert.Empty(t, refusal)

	status, refusal = deployVerdict("regressed", false, nil)
	assert.Equal(t, "regressed", status)
	assert.Empty(t, refusal)
}

func TestDeployVerdict_deployedRefusesWhenTheCheckCommandFails(t *testing.T) {
	t.Parallel()

	status, refusal := deployVerdict("deployed", true, errors.New("exit status 1"))
	assert.Empty(t, status, "a failing check must never append a status")
	assert.NotEmpty(t, refusal, "the refusal explains why nothing was appended")
}

func TestDeployVerdict_deployedProceedsWhenTheCheckCommandPasses(t *testing.T) {
	t.Parallel()

	status, refusal := deployVerdict("deployed", true, nil)
	assert.Equal(t, "deployed", status)
	assert.Empty(t, refusal)
}

func TestDeployVerdict_regressedRefusesWhenTheCheckCommandStillPasses(t *testing.T) {
	t.Parallel()

	status, refusal := deployVerdict("regressed", true, nil)
	assert.Empty(t, status, "a still-passing check contradicts an asserted regression — refuse, don't fabricate")
	assert.NotEmpty(t, refusal)
}

func TestDeployVerdict_regressedProceedsWhenTheCheckCommandFails(t *testing.T) {
	t.Parallel()

	status, refusal := deployVerdict("regressed", true, errors.New("exit status 1"))
	assert.Equal(t, "regressed", status)
	assert.Empty(t, refusal)
}
