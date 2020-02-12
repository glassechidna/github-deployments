package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
	"os"
	"strconv"
	"strings"
)

func main() {
	ownerRepo := os.Getenv("GITHUB_REPOSITORY")
	bits := strings.Split(ownerRepo, "/")
	owner := bits[0]
	repo := bits[1]

	deploymentId, _ := strconv.Atoi(os.Getenv("INPUT_DEPLOYMENTID"))
	if deploymentId == 0 {
		deploymentId, _ = strconv.Atoi(os.Getenv("JOB_DEPLOYMENTID"))
	}

	environment := os.Getenv("INPUT_ENVIRONMENT")
	environmentUrl := os.Getenv("INPUT_ENVIRONMENTURL")
	description := os.Getenv("INPUT_DESCRIPTION")
	token := os.Getenv("INPUT_TOKEN")
	commit := os.Getenv("INPUT_SHA")
	runId := os.Getenv("GITHUB_RUN_ID")
	state := os.Getenv("INPUT_STATE")

	ctx := context.Background()
	st := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	c := github.NewClient(oauth2.NewClient(ctx, st))

	switch state {
	case "create":
		create(ctx, c, owner, repo, commit, environment, description)
	case "in_progress", "error", "failure", "inactive", "queued", "pending", "success":
		update(ctx, c, owner, repo, state, description, environmentUrl, runId, deploymentId)
	}
}

func update(ctx context.Context, c *github.Client, owner, repo, state, description, envUrl, runId string, deploymentId int) {
	logUrl := fmt.Sprintf("https://github.com/%s/%s/actions/runs/%s", owner, repo, runId)
	input := &github.DeploymentStatusRequest{
		State:          &state,
		LogURL:         &logUrl,
		Description:    &description,
		EnvironmentURL: &envUrl,
		AutoInactive:   github.Bool(true),
	}

	if len(description) > 0 {
		input.Description = &description
	}

	if len(envUrl) > 0 {
		input.EnvironmentURL = &envUrl
	}

	_, _, err := c.Repositories.CreateDeploymentStatus(ctx, owner, repo, int64(deploymentId), input)
	if err != nil {
		panic(err)
	}
}

func create(ctx context.Context, c *github.Client, owner, repo, commit, environment, description string) {
	required := []string{}

	input := &github.DeploymentRequest{
		Ref:              &commit,
		Environment:      &environment,
		RequiredContexts: &required,
	}

	if len(description) > 0 {
		input.Description = &description
	}

	deployment, _, err := c.Repositories.CreateDeployment(ctx, owner, repo, input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("::set-output name=deploymentId::%d\n", deployment.GetID())
	fmt.Printf("::set-output name=JOB_DEPLOYMENTID::%d\n", deployment.GetID())
}
