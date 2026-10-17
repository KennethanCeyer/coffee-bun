package checker

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func (c *Checker) detectBinaryVersion(bin string) string {
	switch bin {
	case "go":
		out, err := exec.Command("go", "version").Output()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 3 {
				return fields[2]
			}
		}
	case "node":
		out, err := exec.Command("node", "--version").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	case "python3", "python":
		out, err := exec.Command("python3", "--version").Output()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return strings.TrimSpace(string(out))
		}
		out, err = exec.Command("python", "--version").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	case "java":
		out, err := exec.Command("java", "-version").CombinedOutput()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
		}
	case "gcc", "make":
		out, err := exec.Command(bin, "--version").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
		}
	case "git":
		out, err := exec.Command("git", "--version").Output()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 3 {
				return fields[2]
			}
		}
	case "winget":
		out, err := exec.Command("winget", "--version").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	case "aws":
		out, err := exec.Command("aws", "--version").CombinedOutput()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 1 {
				return fields[0]
			}
		}
	case "gcloud":
		out, err := exec.Command("gcloud", "--version").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
		}
	case "az":
		out, err := exec.Command("az", "--version").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
		}
	case "tree":
		out, err := exec.Command("tree", "--version").Output()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	case "zoxide":
		out, err := exec.Command("zoxide", "--version").Output()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	case "htop":
		out, err := exec.Command("htop", "--version").Output()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	case "fzf":
		out, err := exec.Command("fzf", "--version").Output()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 1 {
				return fields[0]
			}
		}
	case "rg":
		out, err := exec.Command("rg", "--version").Output()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	case "eza":
		out, err := exec.Command("eza", "--version").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "v") {
					parts := strings.Fields(line)
					if len(parts) > 0 {
						return parts[0]
					}
				}
			}
		}
	case "jq":
		out, err := exec.Command("jq", "--version").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	case "bat", "batcat":
		out, err := exec.Command(bin, "--version").Output()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	case "code":
		out, err := exec.Command("code", "--version").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 && lines[0] != "" {
				return strings.TrimSpace(lines[0])
			}
		}
	default:
		out, err := exec.Command(bin, "--version").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
		}
	}
	return ""
}

func (c *Checker) detectAppVersion(appPath string) string {
	plist := filepath.Join(appPath, "Contents", "Info.plist")
	if _, err := os.Stat(plist); err == nil {
		out, err := exec.Command("defaults", "read", plist, "CFBundleShortVersionString").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}

type brewOutdatedResp struct {
	Formulae []struct {
		Name           string `json:"name"`
		CurrentVersion string `json:"current_version"`
	} `json:"formulae"`
	Casks []struct {
		Name           string `json:"name"`
		CurrentVersion string `json:"current_version"`
	} `json:"casks"`
}

// CheckOutdated compares installed packages against brew outdated
func (c *Checker) CheckOutdated(pkgs []model.Package) {
	if !c.IsBrewInstalled() {
		return
	}

	cmd := exec.Command("brew", "outdated", "--json=v2")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return
	}

	c.applyOutdatedJSON(pkgs, stdout.Bytes())
}

func (c *Checker) applyOutdatedJSON(pkgs []model.Package, data []byte) {
	var resp brewOutdatedResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}

	formulaMap := make(map[string]string)
	for _, f := range resp.Formulae {
		formulaMap[f.Name] = f.CurrentVersion
	}

	caskMap := make(map[string]string)
	for _, k := range resp.Casks {
		caskMap[k.Name] = k.CurrentVersion
	}

	for i := range pkgs {
		if pkgs[i].BrewName != "" {
			if pkgs[i].IsCask {
				if latest, ok := caskMap[pkgs[i].BrewName]; ok {
					pkgs[i].Outdated = true
					pkgs[i].LatestVersion = latest
				}
			} else {
				if latest, ok := formulaMap[pkgs[i].BrewName]; ok {
					pkgs[i].Outdated = true
					pkgs[i].LatestVersion = latest
				}
			}
		}
	}
}
