<p align="center">
  <h1 align="center">⚒️ DevForge</h1>
  <p align="center">Elite Development Environment Automation CLI</p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/go-1.22+-00ADD8?style=flat-square&logo=go" alt="Go Version" />
  <img src="https://img.shields.io/badge/platform-macOS%20|%20Linux%20|%20Windows-lightgrey?style=flat-square" alt="Platform" />
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License" />
</p>

---

## What is DevForge?

DevForge is a standalone, heavy-duty CLI tool built for modern developers who value speed, consistency, and safety. It automates your environment scaffolding so you never have to manually install toolchains or configure boilerplate again.

**Not a toy script.** DevForge provides enterprise-level reliability:
- Smart dependency detection and version pinning
- Parallel installations with robust fail-fast limits
- LIFO undo rollback engine (if it fails, it cleans up)
- Safe execution array wrappers (no `sh -c` injections)
- 24-hour template registry caching for immediate response
- Beautiful, intuitive CLI/UX

---

## Installation

### From Source

```bash
git clone https://github.com/ChinmayyK/DevForge.git
cd DevForge
make build          # builds the devforge binary
make build-all      # cross-compiles for macOS/Linux/Windows
```

---

## Quick Start

Initialize a project using a configuration file:

```bash
devforge init my-awesome-app
```

Safety checks included: DevForge will not overwrite `my-awesome-app` without the `--force` flag.

### Check System Health

Check if your local machine has the tools it needs:

```bash
$ devforge doctor

DevForge Doctor — System Check
--------------------------------------------------
OS: macOS/arm64
Package Manager: brew

node       : ✔ installed (v25.3.0)
git        : ✔ installed (git version 2.50.1 (Apple Git-155))
docker     : ✖ missing
--------------------------------------------------
Status: PARTIALLY READY
```

### Dry-Run Mode

Test what DevForge *would* do without modifying your system:

```bash
devforge init test-app --dry-run --verbose
```

---

## Core Commands

| Command | Description |
|---------|-------------|
| `init <name>` | Scaffold a project — launches interactive wizard if no config found |
| `doctor` | Check installed dependencies against your config |
| `templates list`| Browse available starter templates from the remote registry |
| `templates search`| Search the registry by keyword |
| `plugin list` | Discover available local plugins |
| `plugin run` | Execute a plugin with JSON stdin/stdout integration |
| `update` | Auto-update DevForge to the latest GitHub release |
| `completion` | Generate shell completions (bash/zsh/fish/powershell) |
| `version` | Display the binary version, OS context, and Go version |

---

## UX and Error Philosophy

DevForge never panics. All errors are structured to be actionable.

```bash
[ERR_PATH_EXISTS] directory "my-app" already exists
Hint: use the --force flag to overwrite
```

The system output is quiet when it works and descriptive when it fails, utilizing standard ANSI colors (`✔` and `✖`) without noisy stack traces.

---

## Performance & Caching

DevForge is built for speed:
- **Parallel processing:** Installs `node`, `git`, `docker`, etc. concurrently.
- **24-hour Cache:** The template registry lists are cached locally (`~/.devforge/cache/templates.json`) out-of-the-box. Use `--refresh` to bypass.
- **Semantic Versioning Constraints:** Smart version checks quickly determine if `node@18` is already available or needs to be downloaded.

---

## Configuration (`default.yaml`)

```yaml
dependencies:
  - name: node
    version: "20"
  - name: git
  - name: docker

template: https://github.com/ChinmayyK/some-template.git

linting: true
gitHooks: true
envFile: true
```

---

## License

MIT License. Designed for professional engineers.
