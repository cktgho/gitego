# gitego

A context-aware identity manager for Git.

## What is gitego?

Have you ever accidentally committed to a personal project with your work email? Or pushed to a client's repository with your personal GitHub account? `gitego` is a command-line tool designed to completely eliminate that problem.

It acts as a seamless manager for your Git "alter egos," allowing you to define, switch between, and automatically apply different user profiles (`user.name`, `user.email`), SSH keys, and Personal Access Tokens (PATs) depending on your context.

The ambition is to create a single, reliable tool that safeguards your identity across all your projects—work, personal, clients, and open source—making identity management an invisible, automated part of your workflow rather than a manual checklist item.

## Why the name?

The name `gitego` works on two layers, reflecting the project's function and its construction:

* **Git + Ego:** The core concept is managing your different "alter egos" within Git. `gitego` helps you switch between your various digital personas effortlessly.
* **Go:** The name's suffix is a nod to **Go**, the language this tool is written in, known for creating fast, reliable, and cross-platform command-line applications.

## How is it different?

While other solutions exist, `gitego` aims to be more holistic and automated.

| Method | The Problem | How gitego is Better |
| :--- | :--- | :--- |
| **Manual `git config`** | Requires manually setting configs per project. It's repetitive, error-prone, and doesn't scale. | `gitego` provides a clean CLI to manage profiles and apply them with a single command. |
| **Git's `includeIf`** | Powerful, but the syntax is cryptic and requires manually editing config files. It also doesn't handle SSH keys or API tokens. | `gitego` will act as a friendly manager for `includeIf` directives, while also extending the context-switching to **SSH keys and PATs**. |
| **Other Profile Switchers** | Most existing tools only swap the `user.name` and `user.email` in your global gitconfig. | `gitego` manages the entire identity stack: **user info, transport security (SSH), and API authentication (PATs)** in one unified tool. |

## Project roadmap

This project will be developed in phases. Here is the planned feature set:

### Phase 1: Core functionality

-   [ ] Profile management (`add`, `edit`, `rm`, `list`, `use`)
-   [ ] Global profile switching
-   [ ] Repository-specific (`--local`) profile switching
-   [ ] Directory-based automatic profile switching (using `includeIf`)
-   [ ] `gitego status` command to show the currently active profile

### Phase 2: Advanced integrations

-   [ ] Automatic SSH key switching for different profiles/hosts
-   [ ] Secure Personal Access Token (PAT) management via OS keychain
-   [ ] Platform-awareness (GitHub, GitLab, Bitbucket) for PAT storage
-   [ ] Pre-commit hook to warn a user if they are about to commit with a mismatched profile

### Phase 3: Ecosystem and polish

-   [ ] Automatic token validation and expiration monitoring
-   [ ] `gitego check` command to validate the current setup
-   [ ] Support for team-based configuration templates
-   [ ] Import/export functionality for easy sharing of profile templates

## Installation

*(Coming soon...)*

## Basic usage

*(Coming soon, but here is a preview of the intended commands.)*

```bash
# Add a new profile for work
gitego add work --name "Jane Doe" --email "jane.doe@company.com"

# Set a directory to automatically use this profile
gitego auto ~/work/* work

# Check your current status
gitego status
#=> Current: work (jane.doe@company.com)
#=> Rule: Active rule for path ~/work/*
```

## Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.

## License

This project is licensed under the terms of the [MIT License](LICENSE).