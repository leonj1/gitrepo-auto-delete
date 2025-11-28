# ghautodelete

Enable automatic branch deletion on GitHub repositories after PR merges.

```bash
ghautodelete octocat/hello-world
```

That's it. Feature branches will now be automatically deleted when PRs are merged.

## Quick Start (3 minutes)

### 1. Install

```bash
go install github.com/josejulio/ghautodelete/cmd/ghautodelete@latest
```

Or download from [releases](https://github.com/josejulio/ghautodelete/releases).

### 2. Set up authentication

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

Or use the `gh` CLI (token is read automatically from `~/.config/gh/hosts.yml`).

### 3. Enable auto-delete on your repo

```bash
ghautodelete owner/repo
```

Done! Your repository now auto-deletes branches after PR merges.

## Usage

```bash
# Using owner/repo format
ghautodelete octocat/hello-world

# Using HTTPS URL
ghautodelete https://github.com/octocat/hello-world

# Using SSH URL
ghautodelete git@github.com:octocat/hello-world.git

# Check current status without making changes
ghautodelete --check owner/repo

# Preview what would happen
ghautodelete --dry-run owner/repo

# Use explicit token
ghautodelete --token ghp_xxxx owner/repo

# Verbose output
ghautodelete --verbose owner/repo
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Authentication failed |
| 4 | Insufficient permissions |
| 5 | Repository not found |
| 6 | API rate limited |

## Token Requirements

Your GitHub token needs the `repo` scope. Create one at [GitHub Settings > Developer settings > Personal access tokens](https://github.com/settings/tokens).

Token sources (in order of precedence):
1. `--token` flag
2. `GITHUB_TOKEN` environment variable
3. `gh` CLI configuration (`~/.config/gh/hosts.yml`)

## Development

```bash
# Run tests
make test

# Build binary
make build

# Run linter
make lint
```

## License

MIT
