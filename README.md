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

1. **gitlab**
    - **host**: The URL of the GitLab instance you're working with.
    - **token**: The personal access token to authenticate with GitLab.
  
2. **jira**
    - **host**: The URL of your Jira instance.
    - **email**: The email account associated with the Jira instance.
    - **token**: The API token to authenticate with Jira.
  
3. **project**
    - **jira**: Project-specific settings for Jira.
        - **name**: The name of the Jira project.
        - **jql**: Jira Query Language expression for issue filtering.
        - **custom_field**: Custom fields like `story_point` and `epic_start_date`.
    - **gitlab**: Project-specific settings for GitLab.
        - **issue**: Path to the GitLab project where issues will be migrated.
        - **epic**: Path to the GitLab project where epics will be migrated.

```yaml
# Example config.yaml
gitlab:
  host: https://gitlab.com
  token: private-token
...
```

### user.csv

This CSV file contains the mapping of Jira accounts to GitLab accounts. CSV files are excellent for tabular data and are widely supported.

#### **Columns**

1. **Jira Account ID**: The unique identifier for a Jira account.
2. **Jira Display Name**: The display name in Jira.
3. **GitLab User ID**: The unique identifier for a GitLab account.

```csv
# Example user.csv
Jira Account ID, Jira Display Name, GitLab User ID
12372034567899abcde,Seonghun Son,1231231234
...
```
## Usage
<!-- TODO -->
### To start using j2lab
```
brew tap infograb/j2lab
brew install j2lab
```
### To start developing j2lab
<!-- TODO 프로젝트 구조, 코드 설명 -->
## Contribution
If you're interested in contributing, please refer to the [Contributing Guide](./CONTRIBUTING.md) before submitting a pull request.
## Support
For any inquiries or additional questions, please reach out to **InfoGrab** via [email](support@infograb.net) or GitLab/GitHub messages.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

