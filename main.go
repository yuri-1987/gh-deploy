package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
)

var options Options

type Options struct {
	Repo    string `short:"r" long:"repo" description:"Github repository name including owner" required:"true"`
	Env     string `short:"e" long:"env" description:"Github environment name" required:"true"`
	RunId   string `short:"i" long:"run-id" description:"Github Action run id to approve" required:"true"`
	Approve bool   `long:"approve" description:"Approve deployment"`
	Reject  bool   `long:"reject" description:"Reject deployment"`
}



var parser = flags.NewParser(&options, flags.Default)

func validateCliFlags() {
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}

	if options.Approve == options.Reject {
		color.Red("--approve and --reject are mutually exclusive")
		os.Exit(1)
	}
}

func ackDeployment(client api.RESTClient, repo string, runId string, envId int, state string, user string) error {

	type AckDeployment struct {
		EnvironmentIds []int  `json:"environment_ids"`
		State          string `json:"state"`
		Comment        string `json:"comment"`
	}

	body := AckDeployment{
		EnvironmentIds: []int{envId},
		State:          state,
		Comment:        fmt.Sprintf("Deployment %s by %s", state, user),
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	response := []struct {
		Sha         string `json:"sha"`
		Ref         string `json:"ref"`
		Task        string `json:"task"`
		Environment string `json:"environment"`
	}{}
	err = client.Post(fmt.Sprintf("repos/%s/actions/runs/%s/pending_deployments", repo, runId), bytes.NewReader(bodyBytes), &response)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf(
		"Deployment to %s environment was %s by %s\n"+
			"Task: %s\n"+
			"Branch: %s\n",
		green(strings.ToUpper(response[0].Environment)), green(state), green(user), green(response[0].Task), green(response[0].Ref),
	)
	return nil
}

func getEnvIdFromCheckRun(client api.RESTClient, repo string, runId string, env string) (int, error) {

	type PendingApproval struct {
		Environment struct {
			EnvName string `json:"name"`
			EnvId   int    `json:"id"`
		} `json:"environment"`
	}

	var response []PendingApproval

	err := client.Get(fmt.Sprintf("repos/%s/actions/runs/%s/pending_deployments", repo, runId), &response)
	if err != nil {
		return 0, err
	}

	if len(response) == 0 {
		color.Red("No pending deployments found for check run: %s, env: %s\n", runId, env)
		os.Exit(1)
	}

	for i := range response {
		switch response[i].Environment.EnvName {
		case env:
			return response[i].Environment.EnvId, nil
		}
	}

	color.Red(
		`--env|-e "%s" does not match the environment of the pending deployment.`+"\n",
		env,
	)

	return 0, nil
}

func main() {

	validateCliFlags()

	client, err := gh.RESTClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	userName := struct{ Login string }{}
	err = client.Get("user", &userName)
	if err != nil {
		fmt.Println(err)
	}

	envId, err := getEnvIdFromCheckRun(client, options.Repo, options.RunId, options.Env)
	if err != nil {
		fmt.Println(err)
	}

	state := "approved"
	if options.Reject {
		state = "rejected"
	}

	err = ackDeployment(client, options.Repo, options.RunId, envId, state, userName.Login)
	if err != nil {
		fmt.Println(err)
	}
}
