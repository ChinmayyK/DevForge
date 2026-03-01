<p align="center">
  <h1 align="center">⚒️ DevForge</h1>
  <p align="center">Production-grade CLI for automated project scaffolding</p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/go-1.26+-00ADD8?style=flat-square&logo=go" alt="Go Version" />
  <img src="https://img.shields.io/badge/platform-macOS-lightgrey?style=flat-square&logo=apple" alt="Platform" />
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License" />
</p>

---

## What is DevForge?

DevForge is a cross-platform CLI tool that eliminates the tedium of project setup. Point it at a config file and a template repository, and it will:

- **Detect your OS** and validate platform compatibility
- **Install dependencies** via your platform's package manager (Homebrew on macOS)
- **Clone starter templates** from any Git repository
- **Generate `.env` files** from templates with interactive prompts
- **Roll back automatically** if any step fails mid-process
- **Run system health checks** with the `doctor` command

Built with Go for single-binary distribution — no runtime dependencies required.

---

## Installation

### From Source

```bash
git clone https://github.com/chinmay/devforge.git
cd devforge
go build -o devforge .
sudo mv devforge /usr/local/bin/
```

### Verify

```bash
devforge --help
devforge doctor
```

---

## Usage

### Scaffold a New Project

```bash
devforge init my-app
```

This will:
1. Detect your OS and architecture
2. Load configuration from `config/default.yaml`
3. Install any missing dependencies via Homebrew
4. Clone the configured starter template
5. Generate a `.env` file from `.env.template` (if present)
6. Print a success summary with next steps

### Dry Run

Preview what would happen without making any changes:

```bash
devforge init my-app --dry-run
```

### Verbose Logging

Enable debug-level output for troubleshooting:

```bash
devforge init my-app --verbose
```

### Custom Config

Point to a custom configuration file:

```bash
devforge init my-app --config ./myconfig.yaml
```

### System Health Check

```bash
devforge doctor
```

Outputs a formatted table showing the status of required tools:

```
  Tool         Status       Version
  ────         ──────       ───────
  Homebrew     ✓ installed  Homebrew 5.0.15
  Node.js      ✓ installed  v25.3.0
  Git          ✓ installed  git version 2.50.1
  Docker       ✓ installed  Docker version 29.1.3

  ✅ All checks passed — system is ready!
```

---

## Configuration

DevForge uses YAML configuration. Default config is at `config/default.yaml`:

```yaml
dependencies:
  - name: node
  - name: git
  - name: docker

template: "https://github.com/some-org/node-template"

linting: true
gitHooks: true
envFile: true
```

| Field          | Type       | Description                                    |
|----------------|------------|------------------------------------------------|
| `dependencies` | `[]object` | Tools to install before scaffolding             |
| `template`     | `string`   | Git URL of the starter template repository      |
| `linting`      | `bool`     | Enable linting configuration                    |
| `gitHooks`     | `bool`     | Enable git hooks setup                          |
| `envFile`      | `bool`     | Generate `.env` from `.env.template`            |

---

## Architecture

DevForge follows Go best practices with a modular, layered architecture:

```
main.go                    → Entry point
cmd/                       → CLI command definitions (Cobra)
  root.go                  → Root command + global flags
  init.go                  → Project scaffolding orchestration
  doctor.go                → System health checks
internal/                  → Private application packages
  config/config.go         → YAML config loading (Viper)
  logger/logger.go         → Structured logging (logrus)
  osdetect/osdetect.go     → OS detection & validation
  executor/executor.go     → Safe command execution wrapper
  rollback/rollback.go     → LIFO rollback engine
  installer/installer.go   → Package manager interface
  installer/brew.go        → Homebrew implementation
  template/clone.go        → Git template cloning (go-git)
  envgen/envgen.go         → .env file generation
config/                    → Configuration files
  default.yaml             → Default configuration
```

### Key Design Decisions

| Concern               | Approach                                                    |
|------------------------|-------------------------------------------------------------|
| CLI Framework          | Cobra for commands, flags, and help generation               |
| Configuration          | Viper for YAML parsing with validation                       |
| Logging                | logrus with file + console output, verbose toggle            |
| Command Execution      | Custom wrapper with dry-run, input sanitization, structured results |
| Rollback               | LIFO action stack; all critical ops register undo actions    |
| Dependency Injection   | Interfaces for installers; factory function selects platform |
| Error Handling         | Explicit error returns; no panics; wrapped errors throughout |

---

## Folder Structure

```
devforge/
├── cmd/
│   ├── root.go          # Root Cobra command with --config, --dry-run, --verbose
│   ├── init.go          # Init command: full scaffolding orchestration
│   └── doctor.go        # Doctor command: system readiness checks
├── internal/
│   ├── config/
│   │   └── config.go    # Viper-based config loading and validation
│   ├── logger/
│   │   └── logger.go    # logrus logger with file hook
│   ├── osdetect/
│   │   └── osdetect.go  # runtime.GOOS detection
│   ├── executor/
│   │   └── executor.go  # os/exec wrapper with sanitization
│   ├── rollback/
│   │   └── rollback.go  # Reverse-order rollback manager
│   ├── installer/
│   │   ├── installer.go # Installer interface + factory
│   │   └── brew.go      # Homebrew implementation
│   ├── template/
│   │   └── clone.go     # go-git template cloner
│   └── envgen/
│       └── envgen.go    # .env file generator
├── config/
│   └── default.yaml     # Default configuration
├── main.go              # Entry point
├── go.mod
└── go.sum
```

---

## Roadmap

- [ ] **Linux support** — apt/dnf installer implementations
- [ ] **Windows support** — winget/choco installer implementations
- [ ] **Plugin system** — user-defined post-scaffold hooks
- [ ] **Multiple templates** — template marketplace with selection
- [ ] **Interactive mode** — TUI-based guided project setup
- [ ] **CI/CD generation** — GitHub Actions / GitLab CI templates
- [ ] **Monorepo support** — multi-service project scaffolding
- [ ] **Update command** — `devforge update` to upgrade project dependencies
- [ ] **Config validation CLI** — `devforge validate` to lint config files

---

## License

MIT

---

<p align="center">
  Built with ❤️ in Go
</p>
