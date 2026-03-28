package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const configFile = ".irdocker.json"
const version = "1.1.0"

type Config struct {
	Registries []Registry `json:"registries"`
}

type Registry struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

type CheckStatus int

const (
	StatusFound    CheckStatus = iota
	StatusNotFound             // registry reachable, image definitely not there
	StatusTimeout
	StatusNetError
	StatusUnknown // weird response we can't interpret
)

type CheckResult struct {
	Registry Registry
	Status   CheckStatus
	Detail   string // e.g. "HTTP 403" or "DNS lookup failed"
}

var defaultRegistries = []Registry{
	{Name: "ArvanCloud", Host: "docker.arvancloud.ir"},
	{Name: "MobinHost", Host: "docker.mobinhost.com"},
	{Name: "Pardisco", Host: "mirrors.pardisco.co"},
	{Name: "Focker.ir", Host: "focker.ir"},
	{Name: "Kernel.ir", Host: "docker.kernel.ir"},
	{Name: "Megan.ir", Host: "hub.megan.ir"},
	{Name: "Atlantiscloud.ir", Host: "hub.atlantiscloud.ir"},
	{Name: "Iranserver.com", Host: "docker.iranserver.com"},
	{Name: "Liara.ir", Host: "docker-mirror.liara.ir"},
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// ── config ────────────────────────────────────────────────────────────────────

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return configFile
	}
	return filepath.Join(home, configFile)
}

func loadConfig() Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return Config{Registries: defaultRegistries}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{Registries: defaultRegistries}
	}
	if len(cfg.Registries) == 0 {
		cfg.Registries = defaultRegistries
	}
	return cfg
}

func saveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}

// ── image parsing ─────────────────────────────────────────────────────────────

func parseImage(image string) (namespace, name, tag string) {
	tag = "latest"
	if idx := strings.LastIndex(image, ":"); idx != -1 {
		tag = image[idx+1:]
		image = image[:idx]
	}
	parts := strings.SplitN(image, "/", 2)
	if len(parts) == 1 {
		return "library", parts[0], tag
	}
	return parts[0], parts[1], tag
}

// ── auth helpers ──────────────────────────────────────────────────────────────

// parseWWWAuthenticate parses:
//
//	Bearer realm="https://...",service="...",scope="..."
func parseWWWAuthenticate(header string) (realm, service, scope string) {
	header = strings.TrimPrefix(header, "Bearer ")
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(strings.TrimSpace(kv[1]), `"`)
		switch key {
		case "realm":
			realm = val
		case "service":
			service = val
		case "scope":
			scope = val
		}
	}
	return
}

func getToken(realm, service, scope string) (string, error) {
	url := fmt.Sprintf("%s?service=%s&scope=%s", realm, service, scope)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("token parse error: %w", err)
	}
	if result.Token != "" {
		return result.Token, nil
	}
	return result.AccessToken, nil
}

// ── manifest check ────────────────────────────────────────────────────────────

// checkManifest does GET on a manifest URL, handles Bearer auth challenge.
// Returns (httpStatusCode, error).
func checkManifest(url string) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept",
		"application/vnd.docker.distribution.manifest.v2+json, "+
			"application/vnd.docker.distribution.manifest.v1+json, "+
			"application/vnd.oci.image.manifest.v1+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return 200, nil
	}

	if resp.StatusCode == 401 {
		wwwAuth := resp.Header.Get("Www-Authenticate")
		if !strings.HasPrefix(wwwAuth, "Bearer ") {
			// Basic auth — we can't handle this, report as unknown
			return 401, fmt.Errorf("requires Basic authentication")
		}

		realm, service, scope := parseWWWAuthenticate(wwwAuth)
		if realm == "" {
			return 401, fmt.Errorf("malformed Www-Authenticate header")
		}

		token, err := getToken(realm, service, scope)
		if err != nil {
			return 401, fmt.Errorf("token fetch failed: %w", err)
		}

		req2, _ := http.NewRequest("GET", url, nil)
		req2.Header.Set("Accept", req.Header.Get("Accept"))
		req2.Header.Set("Authorization", "Bearer "+token)

		resp2, err := httpClient.Do(req2)
		if err != nil {
			return 0, err
		}
		defer resp2.Body.Close()
		return resp2.StatusCode, nil
	}

	return resp.StatusCode, nil
}

// simplifyError makes net errors readable.
func simplifyError(err error) string {
	s := err.Error()
	switch {
	case strings.Contains(s, "no such host") || strings.Contains(s, "Name or service not known"):
		return "DNS lookup failed (host not found)"
	case strings.Contains(s, "connection refused"):
		return "connection refused"
	case strings.Contains(s, "certificate") || strings.Contains(s, "tls") || strings.Contains(s, "x509"):
		return "TLS/certificate error"
	case strings.Contains(s, "i/o timeout") || strings.Contains(s, "deadline exceeded") || strings.Contains(s, "timeout"):
		return "connection timed out"
	case strings.Contains(s, "EOF"):
		return "connection closed unexpectedly"
	default:
		if idx := strings.LastIndex(s, ": "); idx != -1 {
			return s[idx+2:]
		}
		return s
	}
}

func isTimeout(err error) bool {
	s := err.Error()
	return strings.Contains(s, "deadline exceeded") ||
		strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "timeout")
}

// ── registry check ────────────────────────────────────────────────────────────

func checkRegistry(reg Registry, namespace, name, tag string) CheckResult {
	// Build candidate manifest URLs.
	// For official images (library/nginx), mirrors may or may not include the "library/" prefix.
	urls := []string{}
	if namespace == "library" {
		urls = append(urls,
			fmt.Sprintf("https://%s/v2/library/%s/manifests/%s", reg.Host, name, tag),
			fmt.Sprintf("https://%s/v2/%s/manifests/%s", reg.Host, name, tag),
		)
	} else {
		urls = append(urls,
			fmt.Sprintf("https://%s/v2/%s/%s/manifests/%s", reg.Host, namespace, name, tag),
		)
	}

	var lastErr error
	var lastStatus int

	for _, url := range urls {
		status, err := checkManifest(url)
		if err != nil {
			if isTimeout(err) {
				return CheckResult{reg, StatusTimeout, "connection timed out"}
			}
			lastErr = err
			continue
		}

		switch status {
		case 200:
			return CheckResult{reg, StatusFound, ""}
		case 404:
			// Definitely not there — don't try other URL variants
			return CheckResult{reg, StatusNotFound, "HTTP 404"}
		case 401:
			// Auth failed after token exchange — can't confirm
			return CheckResult{reg, StatusUnknown, "auth required (private registry?)"}
		default:
			lastStatus = status
		}
	}

	if lastErr != nil {
		if isTimeout(lastErr) {
			return CheckResult{reg, StatusTimeout, "connection timed out"}
		}
		return CheckResult{reg, StatusNetError, simplifyError(lastErr)}
	}
	if lastStatus != 0 {
		return CheckResult{reg, StatusNotFound, fmt.Sprintf("HTTP %d", lastStatus)}
	}
	return CheckResult{reg, StatusNotFound, "not available"}
}

// ── pull command ──────────────────────────────────────────────────────────────

func pullCommand(reg Registry, namespace, name, tag string) string {
	if namespace == "library" {
		return fmt.Sprintf("docker pull %s/%s:%s", reg.Host, name, tag)
	}
	return fmt.Sprintf("docker pull %s/%s/%s:%s", reg.Host, namespace, name, tag)
}

// ── CLI commands ──────────────────────────────────────────────────────────────

func cmdCheck(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: irdocker check <image[:tag]>")
		os.Exit(1)
	}

	namespace, name, tag := parseImage(args[0])
	cfg := loadConfig()

	fmt.Printf("\n🔍 Checking image: %s/%s:%s\n", namespace, name, tag)
	fmt.Printf("📦 Checking %d registries...\n\n", len(cfg.Registries))

	results := make([]CheckResult, len(cfg.Registries))
	var wg sync.WaitGroup
	for i, reg := range cfg.Registries {
		wg.Add(1)
		go func(i int, reg Registry) {
			defer wg.Done()
			results[i] = checkRegistry(reg, namespace, name, tag)
		}(i, reg)
	}
	wg.Wait()

	found, notFound, errs := 0, 0, 0

	for _, r := range results {
		label := fmt.Sprintf("%-20s", r.Registry.Name)
		switch r.Status {
		case StatusFound:
			found++
			fmt.Printf("✅ %s → FOUND\n", label)
			fmt.Printf("   %s\n\n", pullCommand(r.Registry, namespace, name, tag))
		case StatusNotFound:
			notFound++
			if r.Detail != "" && r.Detail != "HTTP 404" && r.Detail != "not available" {
				fmt.Printf("❌ %s → NOT FOUND    (%s)\n\n", label, r.Detail)
			} else {
				fmt.Printf("❌ %s → NOT FOUND\n\n", label)
			}
		case StatusTimeout:
			errs++
			fmt.Printf("⏱️  %s → TIMEOUT     (%s)\n\n", label, r.Detail)
		case StatusNetError:
			errs++
			fmt.Printf("🔌 %s → NET ERROR   (%s)\n\n", label, r.Detail)
		case StatusUnknown:
			errs++
			fmt.Printf("⚠️  %s → UNKNOWN     (%s)\n\n", label, r.Detail)
		}
	}

	fmt.Println(strings.Repeat("─", 52))
	parts := []string{fmt.Sprintf("%d found", found), fmt.Sprintf("%d not found", notFound)}
	if errs > 0 {
		parts = append(parts, fmt.Sprintf("%d error(s)", errs))
	}
	fmt.Printf("Result: %s\n\n", strings.Join(parts, ", "))
}

func cmdAdd(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: irdocker add <name> <host>")
		os.Exit(1)
	}
	name := args[0]
	host := strings.TrimPrefix(strings.TrimPrefix(args[1], "https://"), "http://")
	host = strings.TrimRight(host, "/")

	cfg := loadConfig()
	for _, r := range cfg.Registries {
		if r.Host == host {
			fmt.Printf("⚠️  Registry '%s' already exists.\n", host)
			return
		}
	}
	cfg.Registries = append(cfg.Registries, Registry{Name: name, Host: host})
	if err := saveConfig(cfg); err != nil {
		fmt.Printf("❌ Failed to save: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Registry '%s' (%s) added.\n", name, host)
}

func cmdRemove(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: irdocker remove <host>")
		os.Exit(1)
	}
	host := args[0]
	cfg := loadConfig()
	var newList []Registry
	removed := false
	for _, r := range cfg.Registries {
		if r.Host == host {
			removed = true
			continue
		}
		newList = append(newList, r)
	}
	if !removed {
		fmt.Printf("⚠️  Registry '%s' not found.\n", host)
		return
	}
	cfg.Registries = newList
	if err := saveConfig(cfg); err != nil {
		fmt.Printf("❌ Failed to save: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Registry '%s' removed.\n", host)
}

func cmdList() {
	cfg := loadConfig()
	fmt.Printf("\n📋 Configured registries (%d):\n\n", len(cfg.Registries))
	for i, r := range cfg.Registries {
		fmt.Printf("  %d. %-20s %s\n", i+1, r.Name, r.Host)
	}
	fmt.Println()
}

func mirroredImageStr(reg Registry, namespace, name, tag string) string {
	if namespace == "library" {
		return fmt.Sprintf("%s/%s:%s", reg.Host, name, tag)
	}
	return fmt.Sprintf("%s/%s/%s:%s", reg.Host, namespace, name, tag)
}

func cmdCompose(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ Cannot read file: %v\n", err)
		os.Exit(1)
	}
	content := string(data)

	imageLineRe := regexp.MustCompile(`(?m)^(\s+image:\s+)(.+)$`)
	allMatches := imageLineRe.FindAllStringSubmatch(content, -1)

	seen := map[string]bool{}
	var images []string
	for _, m := range allMatches {
		img := strings.TrimSpace(m[2])
		img = strings.Trim(img, `'"`)
		if img != "" && !seen[img] {
			seen[img] = true
			images = append(images, img)
		}
	}

	if len(images) == 0 {
		fmt.Println("⚠️  No image entries found in the compose file.")
		os.Exit(0)
	}

	cfg := loadConfig()
	fmt.Printf("\n🐳 Docker Compose: %s\n", filePath)
	fmt.Printf("📦 Found %d unique image(s), checking %d registries...\n\n", len(images), len(cfg.Registries))

	type imgResult struct {
		original string
		mirrored string
		regName  string
		found    bool
	}

	imgResults := make([]imgResult, len(images))
	var wg sync.WaitGroup
	for i, img := range images {
		wg.Add(1)
		go func(i int, img string) {
			defer wg.Done()
			namespace, name, tag := parseImage(img)
			regResults := make([]CheckResult, len(cfg.Registries))
			var wg2 sync.WaitGroup
			for j, reg := range cfg.Registries {
				wg2.Add(1)
				go func(j int, reg Registry) {
					defer wg2.Done()
					regResults[j] = checkRegistry(reg, namespace, name, tag)
				}(j, reg)
			}
			wg2.Wait()
			for _, r := range regResults {
				if r.Status == StatusFound {
					imgResults[i] = imgResult{
						original: img,
						mirrored: mirroredImageStr(r.Registry, namespace, name, tag),
						regName:  r.Registry.Name,
						found:    true,
					}
					return
				}
			}
			imgResults[i] = imgResult{original: img}
		}(i, img)
	}
	wg.Wait()

	// Build replacement map
	replacements := map[string]string{}
	for _, r := range imgResults {
		if r.found {
			replacements[r.original] = r.mirrored
		}
	}

	// Replace image lines in content
	newContent := imageLineRe.ReplaceAllStringFunc(content, func(match string) string {
		sub := imageLineRe.FindStringSubmatch(match)
		prefix := sub[1]
		img := strings.Trim(strings.TrimSpace(sub[2]), `'"`)
		if mirrored, ok := replacements[img]; ok {
			return prefix + mirrored
		}
		return match
	})

	// Output paths
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	oldBase := stem + ".old" + ext
	mirroredPath := filepath.Join(dir, "docker-compose-mirrored.yaml")
	oldPath := filepath.Join(dir, oldBase)

	if err := os.WriteFile(mirroredPath, []byte(newContent), 0644); err != nil {
		fmt.Printf("❌ Failed to write %s: %v\n", mirroredPath, err)
		os.Exit(1)
	}

	// Print table
	maxImgLen := 10
	for _, r := range imgResults {
		if len(r.original) > maxImgLen {
			maxImgLen = len(r.original)
		}
	}
	hdrFmt := fmt.Sprintf("  %%s %%-%ds  %%-%ds  %%s\n", maxImgLen+2, 20)
	fmt.Printf("📋 Image Mirror Report:\n\n")
	fmt.Printf(hdrFmt, " ", "Image", "Registry", "Mirrored Image")
	fmt.Println("  " + strings.Repeat("─", maxImgLen+55))
	for _, r := range imgResults {
		if r.found {
			fmt.Printf(hdrFmt, "✅", r.original, r.regName, r.mirrored)
		} else {
			fmt.Printf(hdrFmt, "❌", r.original, "—", "no mirror found")
		}
	}
	fmt.Println()

	found := 0
	for _, r := range imgResults {
		if r.found {
			found++
		}
	}
	fmt.Printf("  %d/%d images mirrored → wrote %s\n\n", found, len(images), mirroredPath)

	// Print apply commands
	fmt.Printf("🔧 Apply changes:\n\n")
	fmt.Printf("  mv %s %s\n", filePath, oldPath)
	fmt.Printf("  mv %s %s\n", mirroredPath, filePath)
	if dir == "." {
		fmt.Printf("  docker compose up -d\n\n")
	} else {
		fmt.Printf("  docker compose -f %s up -d\n\n", filePath)
	}
}

func cmdReset() {
	cfg := Config{Registries: defaultRegistries}
	if err := saveConfig(cfg); err != nil {
		fmt.Printf("❌ Failed to reset: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Config reset to defaults.")
	cmdList()
}

func usage() {
	fmt.Printf(`
irdocker v%s — Check Iranian Docker Mirrors

USAGE:
  irdocker <image[:tag]>           Check image across all registries
  irdocker check <image[:tag]>     Same as above
  irdocker <compose-file.yaml>     Mirror all images in a docker-compose file
  irdocker list                    List configured registries
  irdocker add <name> <host>       Add a new registry
  irdocker remove <host>           Remove a registry
  irdocker reset                   Reset to default registries
  irdocker help                    Show this help

EXAMPLES:
  irdocker nginx
  irdocker nginx:1.25-alpine
  irdocker gitea/gitea:latest
  irdocker docker-compose.yaml
  irdocker add RunFlare mirror-docker.runflare.com
  irdocker remove focker.ir

STATUS ICONS:
  ✅  Image found — pull command shown below
  ❌  Registry reachable, image not there
  ⏱️   Connection timed out
  🔌  Network error (DNS fail, TLS, refused…)
  ⚠️   Unknown (auth required or unexpected response)

`, version)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(0)
	}

	cmd := args[0]
	switch cmd {
	case "help", "--help", "-h":
		usage()
	case "list", "ls":
		cmdList()
	case "add":
		cmdAdd(args[1:])
	case "remove", "rm":
		cmdRemove(args[1:])
	case "reset":
		cmdReset()
	case "check":
		cmdCheck(args[1:])
	default:
		if strings.HasPrefix(cmd, "-") {
			fmt.Printf("Unknown flag: %s\n", cmd)
			usage()
			os.Exit(1)
		}
		if strings.HasSuffix(cmd, ".yaml") || strings.HasSuffix(cmd, ".yml") {
			cmdCompose(cmd)
		} else {
			cmdCheck(args)
		}
	}
}
