# gh-deploy
A `gh` cli extension to approve or reject pending deployments that are waiting for review.

## Installation

Installation requires a minimum version (2.0.0) of the the Github CLI to support extensions.

1. Install the `gh cli` - see the [installation/upgrade instructions](https://github.com/cli/cli#installation)

2. Install this extension:
```sh
gh extension install yuri-1987/gh-deploy
```

## Usage
```sh
gh deploy --help

Usage:
  gh-deploy [OPTIONS]

Application Options:
  -r, --repo=    Github reposiory name including owner
  -e, --env=     Github environment name
  -i, --run-id=  Github Action run id to approve
      --approve  Approve deployment
      --reject   Reject deployment

Help Options:
  -h, --help     Show this help message
```
To **approve** pending deployment by run-id:
```sh
gh deploy --env prod --run-id 1723641358 --repo "yuri-1987/tf-iac" --approve
```
To **reject** pending deployment by run-id:
```sh
gh deploy --env prod --run-id 1723641358 --repo "yuri-1987/tf-iac" --reject
```

## Use cases
Set environment protection rules for your Github repository, ie. `prod` environment requires a review from your team members.

send them a slack message with the details of the deployment via github action.

**_workflow snippet:_**
```yaml
deployment_job_name:
  runs-on: ubuntu-latest
  steps:
    - name: send slack message for approval
      uses: act10ns/slack@v1.5.0
      with:
        status: "tf prod waiting for approval"
        channel: '#deployments'
        message: |
          *Terraform Apply for Production env in `${{ github.repository }}` is pending approval*
          Review requested by ${{ github.actor }}
          Please review it here: <${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}|${{ github.run_id }}>
          To *approve*, run:
          ```
          gh deploy --env prod --run-id ${{ github.run_id }} --repo "${{ github.repository }}" --approve
          ```
          To *reject*, run:
          ```
          gh deploy --env prod --run-id ${{ github.run_id }} --repo "${{ github.repository }}" --reject
          ```
          <{{diffUrl}}|Link to PR diff in branch: `{{branch}}`>
```


