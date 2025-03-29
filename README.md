# An Opinionated GitHub Classroom CLI

This extension is an opinionated [GitHub Classroom](https://classroom.github.com) extension for GitHub CLI to easily work with GitHub Classrooms and student repos. Its main purpose is to clone GitHub Classroom assignment and the starter repo as well as to synch back changes from the starter repo to the student repos.

# Installation
- Install the gh cli

  On MacOS, e.g., [Homebrew](https://brew.sh/) can be used

  ```bash
  brew install gh
  ```
- Install this extension
  ```bash
  gh extension install majikmate/gh-crm
  ```
- To update this extension
  ```bash
  gh extension upgrade crm
  ```

# Usage

## Initialization

In order to start with the tool and initialize a classroom repository on the local file system, an Excel file containing a list of students with additional metadata is required in the local folder that should become the root of the local classroom repository.

The Excel file needs to contain a header line in the first row containing following fields:
- Name         ... Full name of the student
- Email        ... Email address of the student
- GitHub User  ... GitHub username of the student

The additional lines need to contain at least one line with the respective student information.

The Email should contain Emails of the students in the format
- *firstname*.*lastname*@domain.tld

The Excel file must be named with a prefix of *account* or *Account* and should have the file extension *.xlsx*. This file can be created, e.g., by gathering student details through a [Microsoft Office Forms](http://forms.office.com/) form and exporting the responses.

See [Commands](#commands) for further details.

### Commands

For more information and a list of available commands

```bash
gh crm -h
```

## License

This project is licensed under the terms of the MIT open source license. Please refer to [LICENSE](./LICENSE) for the full terms.

## Maintainers

See [CODEOWNERS](./CODEOWNERS)

## Attribution and Thanks

This extension is heavily inspired by the great original GitHub Classroom CLI available here:

- [GitHub Classroom CLI](https://github.com/github/gh-classroom)

**License**
- [Orignial License 1](./LICENSE-1.txt)
- [Orignial License 2](./LICENSE-2.txt)