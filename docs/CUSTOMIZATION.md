<h1 align="center">coffee-bun Customization & Enterprise Guide</h1>

coffee-bun is architected from the ground up to support seamless organization-specific customization. Whether you need to distribute company-internal CLI tools, configure enterprise VPN profiles, connect to private package registries, or manage standardized developer presets across an engineering department, coffee-bun provides a clean, two-tier customization framework:

1. **Declarative Configuration (Zero Go code)**: Define company tools, categories, and lifecycle hooks in a clean YAML specification (`bun.yaml` or `/etc/coffee-bun/bun.yaml`).
2. **Code-as-Documentation Architecture**: Clean, modular Go domain files where each category is a standalone, self-documenting package specification.

---

## 1. Declarative Configuration (`bun.yaml`)

You can introduce new packages, override existing tools, organize custom categories, and hook into installation lifecycles without compiling Go code.

### Configuration Resolution Order

coffee-bun automatically discovers and merges configurations using the following precedence (highest priority to lowest):

1. **Explicit CLI Flag**: `bun --config /path/to/custom.yaml` or environment variable `BUN_CONFIG=/path/to/custom.yaml`.
2. **Project / Workspace Root**: `./bun.yaml`, `./.bun.yaml`, or `./coffee-bun.yaml` (ideal for team repositories).
3. **User Personal Profile**: `~/.config/coffee-bun/bun.yaml`, `~/.config/coffee-bun/config.yaml`, `~/.bun.yaml`, or `~/.coffee-bun.yaml`.
4. **System / MDM Provisioned**: `/etc/coffee-bun/bun.yaml` or `/etc/coffee-bun/config.yaml` (ideal for IT/MDM deployment via Jamf, Kandji, or Fleet).

---

## 2. Complete YAML Schema Reference

Create a `bun.yaml` file with the following schema:

```yaml
version: "1.0"

# Lifecycle hooks executed before and after the parallel installation engine
hooks:
  pre_install: |
    echo "[onboarding] Verifying corporate network access..."
  post_install: |
    echo "[onboarding] Standardizing git and corporate environment..."
    git config --global user.email "$(whoami)@acme.corp"

# Custom Category Groups in the Tree TUI
groups:
  - category: "Acme Security"
    title: "Corporate Security & VPN"
    description: "Enterprise WireGuard VPN, 1Password, and device certificates"
    icon: "[sec]"
    packages:
      - acme-wireguard
      - 1password-cli
      - acme-cert-check

  - category: "Backend Stack"
    title: "Acme Backend Engineering Stack"
    description: "Docker engine, PostgreSQL client, and Kubernetes toolset"
    icon: "[corp]"
    packages:
      - orbstack
      - kubectl
      - acme-dev-cli

# Package Specifications
packages:
  # 1. Custom Shell Script Package
  - id: acme-wireguard
    name: "Acme WireGuard VPN"
    description: "Corporate VPN tunnel client with single sign-on"
    category: "Acme Security"
    type: "script"
    check_cmd: "scutil --nc status 'AcmeVPN' 2>/dev/null | grep -q 'Connected' && echo 'Connected' || false"
    check_file: "/Library/Application Support/Acme/vpn.conf"
    install_cmd: |
      curl -fsSL https://internal.acme.corp/vpn/setup-mac.sh | bash
    requires_sudo: true
    sudo_snippet: "sudo /Library/Application Support/Acme/vpn-helper --install"

  # 2. Homebrew Formula / Tap
  - id: acme-dev-cli
    name: "Acme Deployer CLI"
    description: "Internal microservices orchestrator and deployment CLI"
    category: "Backend Stack"
    type: "brew_formula"
    brew_name: "acme-corp/tap/acme-cli"
    binary_name: "acme"
    check_cmd: "acme version"

  # 3. Homebrew GUI Application (Cask)
  - id: orbstack
    name: "OrbStack"
    description: "Fast, light, and battery-friendly Docker container runtime"
    category: "Backend Stack"
    type: "brew_cask"
    brew_name: "orbstack"
    is_cask: true
    app_path: "/Applications/OrbStack.app"

  # 4. Git Repository Package
  - id: acme-dotfiles
    name: "Acme Team Dotfiles"
    description: "Standardized team shell aliases and editor configurations"
    category: "Backend Stack"
    type: "git_repo"
    git_url: "https://github.com/acme-corp/developer-dotfiles.git"
    install_path: "~/.config/acme-dotfiles"

  # 5. Overriding a Core Package (e.g. customized Go or Git)
  - id: git
    name: "Enterprise Git"
    description: "Git configured with Acme internal commit signing"
    env:
      GIT_AUTHOR_NAME: "Acme Developer"
```

---

## 3. Package Field Specification

| Field | Type | Description |
|---|---|---|
| `id` | string | Unique identifier for the package. If matches an existing core package ID, it overrides that package. |
| `name` | string | Human-readable name rendered in the TUI, tables, and logs. |
| `description` | string | Functional explanation shown in the tree and inspector views. |
| `category` | string | Category domain name. Custom categories automatically receive group headers. |
| `type` | string | `brew_formula`, `brew_cask`, `script`, `command`, `git_repo`, `font`, `custom`. |
| `brew_name` | string | Formula or cask identifier passed to Homebrew (supports third-party taps). |
| `is_cask` | boolean | Set `true` if the package is a macOS GUI desktop app installed via `brew install --cask`. |
| `binary_name` | string | Name of the executable tested via `PATH` lookup (`exec.LookPath`). |
| `app_path` | string | Path to application bundle (e.g. `/Applications/Slack.app` or `~/Applications/...`). |
| `git_url` | string | Git remote URL cloned during installation. |
| `install_path` | string | Destination directory for cloned git repositories or downloaded resources (supports `~`). |
| `check_cmd` | string | Shell command executed via `bash -c`. If exit code is 0, the first stdout line is captured as `Version`. |
| `check_file` | string | Path to a file or directory whose existence confirms installation (supports `~`). |
| `install_cmd` | string | Shell command or multi-line script executed during installation with live stdout streaming. |
| `update_cmd` | string | Shell command executed during `bun update`. |
| `env` | map | Key-value pairs injected into the environment during installation command execution. |
| `requires_sudo` | boolean | Set `true` if manual user elevation or sudo credentials are required. |
| `sudo_snippet` | string | Exact command snippet displayed to the user when manual intervention is needed. |

---

## 4. Enterprise Recipes

### Recipe A: Corporate Private Package Registries

When engineers need internal npm, pip, and Go module proxy settings configured during laptop setup:

```yaml
version: "1.0"
hooks:
  post_install: |
    # Configure internal npm registry
    npm config set registry https://npm.acme.corp/
    # Configure internal pip index
    mkdir -p ~/.pip
    cat <<EOF > ~/.pip/pip.conf
    [global]
    index-url = https://artifactory.acme.corp/api/pypi/pypi-remote/simple
    EOF
    # Configure GOPROXY
    go env -w GOPROXY="https://goproxy.acme.corp,direct"
    go env -w GONOSUMDB="github.com/acme-corp/*"
```

### Recipe B: Organization-Wide MDM Zero-Touch Provisioning

For IT teams deploying via Jamf Pro, Kandji, or Fleet:

1. Push a configuration profile or script that writes `/etc/coffee-bun/bun.yaml`.
2. Run the unattended installation in the onboarding script:
   ```bash
   bun install all
   ```
3. Verify system health at the end of setup:
   ```bash
   bun doctor
   ```

Because `/etc/coffee-bun/bun.yaml` is located at the system level, all newly onboarded developers on that machine inherit the standardized toolchain automatically.

---

## 5. Code-as-Documentation Architecture

The core codebase of coffee-bun is partitioned into clean, self-documenting domain modules under [`internal/packages/`](file:///Users/pignose/bundle/internal/packages/):

```text
internal/packages/
├── domain_system.go      - System & Toolchains (Xcode Tools, Homebrew, Git)
├── domain_terminal.go    - Terminal & Shell (iTerm2, Fonts, OMZ, p10k, plugins)
├── domain_cloud.go       - Cloud & DevOps CLIs (AWS, Google Cloud, Azure)
├── domain_editors.go     - Editors & Vim Runtime (amix/vimrc, VS Code, code CLI)
├── domain_languages.go   - Compilers & Runtimes (Go, Python 3, Java, GCC, Node)
├── domain_cli.go         - Modern CLI Utilities (zoxide, htop, fzf, ripgrep, eza, tree, jq, bat)
├── domain_apps.go        - Essential macOS Apps (Google Chrome, Default browser)
└── registry.go           - Consolidated registry and customization merger
```

### Uniform Domain Pattern

Each domain file functions as living technical documentation. It declares:
- A descriptive domain header comment.
- A `model.CustomGroup` defining the title, description, badge icon, and members.
- A slice of `model.Package` specifications using the exact same schema.

#### Example: Adding a New Tool in Go

To add a new tool directly to a core domain or an enterprise fork, simply add a package declaration following the standard format:

```go
{
    ID:          "docker",
    Name:        "Docker CLI",
    Description: "Pack, ship, and run distributed applications in containers",
    Category:    model.CategoryCLITools,
    Type:        model.TypeBrewFormula,
    BrewName:    "docker",
    BinaryName:  "docker",
    CheckCmd:    "docker --version",
},
```

No switch statements, hardcoded UI adjustments, or runner plumbing are needed. The checker, runner, updater, table generator, and interactive TUI automatically discover the new tool.
