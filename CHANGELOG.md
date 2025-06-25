# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-06-25

### Added

- **`edit` Command**: New `gitego edit <profile_name>` command allows for modification of existing profiles, including their name, email, username, SSH key, and PAT.
- **`--version` Flag**: Added a `-v` / `--version` flag to the root command to print the application's version number.
- **Shell Completion**: Introduced a `gitego completion [shell]` command to generate auto-completion scripts for Bash, Zsh, Fish, and PowerShell, improving user experience and discoverability.
- **Configuration Validation**: Implemented a validation step on application startup that warns users about inconsistencies in their `config.yaml`, such as auto-switch rules that point to non-existent profiles.

### Changed

- **Enhanced `list` Command**: The `gitego list` command output is now a more informative table, indicating the active global profile with an asterisk (`*`) and showing which profiles have associated SSH keys or PATs.
- **Smarter Hook Installation**: The `gitego install-hook` command now detects existing `pre-commit` hooks. If a hook exists, it will prompt the user for permission to append the `gitego` command instead of failing.
- **Destructive Command Confirmation**: The `gitego rm` command now requires user confirmation before deleting a profile. A `--force` flag was added to allow bypassing this safety check in scripts.
- **Refactored Codebase**:
    - Centralized the logic for determining the active profile based on directory rules into a single `config.GetActiveProfileForCurrentDir()` function, removing duplicated code from the `status` and `check-commit` commands.
    - Consolidated platform-specific keychain logic. Common functions for `gitego`'s internal token vault are now shared, with only OS-specific credential helper logic remaining in separate files.
    - Moved all Git configuration functions (`get` and `set`) into the `utils` package for better code organization.

### Removed

- Removed redundant, local implementations of Git configuration and profile-finding logic from individual command files in favor of centralized helper functions.
