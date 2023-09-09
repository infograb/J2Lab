# j2lab
Jira to GitLab Migrator
## Introduction
j2lab is an open-source project designed to facilitate the seamless migration of projects from Jira to GitLab. j2lab aims to provide a one-stop solution for teams looking to transition their project management tools without losing crucial data.

## Motivation
As of February 2, 2021, Atlassian has ceased selling new server licenses and has announced that they will terminate official support for server licenses in 2024. This change has significant repercussions for organizations that rely on server-based solutions, pushing them to seek alternatives. Faced with the uncertainty brought about by Atlassian's policy change, we have identified GitLab as a robust alternative that offers similar, if not better, functionalities compared to Jira.

## Configuration

### config.yaml

This YAML file houses the configuration for GitLab and Jira connections, as well as project-related settings. YAML files use a human-readable data serialization standard, making them convenient for configuration files.

#### **Structure**
  
1. **jira**
    - **host**: The URL of your Jira instance
    - **name**: The name of the Jira project.
    - **jql**: Jira Query Language expression for issue filtering.
    - **custom_field**: Custom fields like `story_point` and `epic_start_date`.

2. **gitlab**
    - **host**: The URL of the GitLab instance you're working with.
    - **issue**: Path to the GitLab project where issues will be migrated.
    - **epic**: Path to the GitLab project where epics will be migrated.
```yaml
# Example config.yaml
jira:
  host: https://jira.sbx.infograb.io
  name: SSP
  # jql: id = SSP-1029 OR id = SSP-1 
  jql: ID = SSP-25
  custom_field:
    story_point: customfield_10035
    epic_start_date: customfield_10015
    parent_epic: customfield_10110

gitlab:
  host: https://gitlab.com
  issue: infograb/team/devops/toy/gos/poc/jeff
  epic: infograb/team/devops/toy/gos/poc
...
```

### user.csv

This CSV file contains the mapping of Jira username to GitLab accounts. CSV files are excellent for tabular data and are widely supported.

#### **Columns**

1. **Jira Account ID**: The unique identifier for a Jira account.
2. **Jira Display Name**: The display name in Jira.
3. **GitLab User ID**: The unique identifier for a GitLab account.

```csv
# Example user.csv
Jira User Name,GitLab User ID
jeff,1341
Dexter,2155
kane,334
admin,115
...
```
## Installation

### Homebrew (macOS)

```
brew tap infograb/j2lab
brew install j2lab
```
### Executable Files (For other platforms)
You can also download the pre-built binaries for your specific platform from the [Releases](https://gitlab.com/infograb-public/J2Lab/-/releases) page.
<!-- TODO -->
## Usage
```bash
Usage:
  j2lab [flags]
  j2lab [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Modify config files
  help        Help about any command
  run         Run the application
  version     Print the client and server version information

Flags:
  -c, --config string   config.yaml file
  -d, --debug           debug mode
  -h, --help            help for j2lab
  -u, --user string     user.csv file

Use "j2lab [command] --help" for more information about a command.
```
#### example
```bash
export GITLAB_TOKEN=your_gitlab_token
export JIRA_TOKEN=your_jira_token
j2lab run -c config.yaml -u user.csv
```

## Contribution
If you're interested in contributing, please refer to the [Contributing Guide](./CONTRIBUTING.md) before submitting a pull request.
## Support
For any inquiries or additional questions, please reach out to **InfoGrab** via [email](support@infograb.net) or GitLab/GitHub messages.

## License

This project is licensed under the GNU License - see the [LICENSE](./LICENSE) file for details.

