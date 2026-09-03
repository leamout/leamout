package leamout

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/security/token"
)

const (
	deploymentStateSchemaVersion = 1
	deploymentMode               = "self-hosted"
)

type deploymentState struct {
	SchemaVersion int       `json:"schema_version"`
	DeploymentID  string    `json:"deployment_id"`
	Mode          string    `json:"mode"`
	CreatedAt     time.Time `json:"created_at"`
}

type deploymentSecrets struct {
	FreeSWITCHESLPassword           string
	CarrierCredentialEncryptionKey string
	TURNAuthSecret                  string
	PostgresPassword                string
	RedisPassword                   string
	NATSPassword                    string
}

func runInit(stdout, stderr io.Writer) int {
	return runInitAt(stdout, stderr, "/etc/leamout", "/var/lib/leamout", "/var/log/leamout")
}

func runInitAt(stdout, stderr io.Writer, configDir, stateDir, logDir string) int {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		writef(stderr, "unsupported host: %s/%s; Self-Hosted Production v0.1 supports linux/amd64\n", runtime.GOOS, runtime.GOARCH)
		return 1
	}

	if err := ensureDeploymentDirectories(configDir, stateDir, logDir); err != nil {
		writef(stderr, "initialize Leamout directories: %v\n", err)
		return 1
	}

	statePath := filepath.Join(stateDir, "deployment.json")
	envPath := filepath.Join(configDir, "leamout.env")

	stateExists, err := pathExists(statePath)
	if err != nil {
		writef(stderr, "inspect deployment state: %v\n", err)
		return 1
	}
	envExists, err := pathExists(envPath)
	if err != nil {
		writef(stderr, "inspect deployment configuration: %v\n", err)
		return 1
	}

	if stateExists || envExists {
		if !stateExists || !envExists {
			writef(stderr, "incomplete Leamout initialization: %s and %s must either both exist or both be absent\n", statePath, envPath)
			return 1
		}
		state, err := loadDeploymentState(statePath)
		if err != nil {
			writef(stderr, "load deployment identity: %v\n", err)
			return 1
		}
		if err := validateRuntimeEnv(envPath, state.DeploymentID); err != nil {
			writef(stderr, "validate deployment configuration: %v\n", err)
			return 1
		}
		writef(stdout, "✓ Deployment already initialized: %s\n", state.DeploymentID)
		writeln(stdout, "✓ Existing deployment identity and secrets preserved")
		return 0
	}

	state := deploymentState{
		SchemaVersion: deploymentStateSchemaVersion,
		DeploymentID:  uuid.NewString(),
		Mode:          deploymentMode,
		CreatedAt:     time.Now().UTC(),
	}
	secrets, err := generateDeploymentSecrets()
	if err != nil {
		writef(stderr, "generate deployment secrets: %v\n", err)
		return 1
	}

	stateBytes, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		writef(stderr, "encode deployment identity: %v\n", err)
		return 1
	}
	stateBytes = append(stateBytes, '\n')
	envBytes := renderRuntimeEnv(state, secrets)

	if err := writeInitializationFiles(statePath, stateBytes, envPath, envBytes); err != nil {
		writef(stderr, "persist deployment initialization: %v\n", err)
		return 1
	}

	writeln(stdout, "✓ Deployment identity created")
	writeln(stdout, "✓ Deployment-owned secrets generated")
	writeln(stdout, "✓ Production configuration written")
	writef(stdout, "Deployment ID: %s\n", state.DeploymentID)
	writeln(stdout, "Runtime bundle installation, TLS/network validation, and activation remain pending Self-Hosted Production v0.1 work.")
	return 0
}

func ensureDeploymentDirectories(configDir, stateDir, logDir string) error {
	for _, dir := range []string{
		configDir,
		filepath.Join(configDir, "certs"),
		filepath.Join(configDir, "license"),
		stateDir,
		filepath.Join(stateDir, "backups"),
		filepath.Join(stateDir, "data"),
		filepath.Join(stateDir, "data", "postgres"),
		filepath.Join(stateDir, "data", "redis"),
		filepath.Join(stateDir, "data", "nats"),
		filepath.Join(stateDir, "data", "recordings"),
		logDir,
	} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
		if err := os.Chmod(dir, 0o750); err != nil {
			return fmt.Errorf("secure %s: %w", dir, err)
		}
	}
	return nil
}

func generateDeploymentSecrets() (deploymentSecrets, error) {
	freeswitchPassword, err := token.Generate(32)
	if err != nil {
		return deploymentSecrets{}, err
	}
	turnSecret, err := token.Generate(32)
	if err != nil {
		return deploymentSecrets{}, err
	}
	postgresPassword, err := token.Generate(32)
	if err != nil {
		return deploymentSecrets{}, err
	}
	redisPassword, err := token.Generate(32)
	if err != nil {
		return deploymentSecrets{}, err
	}
	natsPassword, err := token.Generate(32)
	if err != nil {
		return deploymentSecrets{}, err
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return deploymentSecrets{}, err
	}
	return deploymentSecrets{
		FreeSWITCHESLPassword:           freeswitchPassword,
		CarrierCredentialEncryptionKey: base64.RawURLEncoding.EncodeToString(key),
		TURNAuthSecret:                  turnSecret,
		PostgresPassword:                postgresPassword,
		RedisPassword:                   redisPassword,
		NATSPassword:                    natsPassword,
	}, nil
}

func renderRuntimeEnv(state deploymentState, secrets deploymentSecrets) []byte {
	return []byte(strings.Join([]string{
		"# Generated by `leamout init`. Do not edit secret values manually.",
		"APP_ENV=production",
		"LEAMOUT_DEPLOYMENT_MODE=" + state.Mode,
		"LEAMOUT_DEPLOYMENT_ID=" + state.DeploymentID,
		"FREESWITCH_ESL_PASSWORD=" + secrets.FreeSWITCHESLPassword,
		"CARRIER_CREDENTIAL_ENCRYPTION_KEY=" + secrets.CarrierCredentialEncryptionKey,
		"TURN_AUTH_SECRET=" + secrets.TURNAuthSecret,
		"POSTGRES_PASSWORD=" + secrets.PostgresPassword,
		"REDIS_PASSWORD=" + secrets.RedisPassword,
		"NATS_PASSWORD=" + secrets.NATSPassword,
		"",
	}, "\n"))
}

func writeInitializationFiles(statePath string, state []byte, envPath string, env []byte) error {
	if err := writeExclusiveFile(statePath, state, 0o600); err != nil {
		return err
	}
	if err := writeExclusiveFile(envPath, env, 0o600); err != nil {
		if removeErr := os.Remove(statePath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("write runtime configuration: %w; rollback deployment state: %w", err, removeErr)
		}
		return err
	}
	return nil
}

func writeExclusiveFile(path string, content []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	ok := false
	defer func() {
		_ = file.Close()
		if !ok {
			_ = os.Remove(path)
		}
	}()
	if _, err := file.Write(content); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	ok = true
	return nil
}

func loadDeploymentState(path string) (deploymentState, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return deploymentState{}, err
	}
	var state deploymentState
	if err := json.Unmarshal(content, &state); err != nil {
		return deploymentState{}, fmt.Errorf("decode %s: %w", path, err)
	}
	if state.SchemaVersion != deploymentStateSchemaVersion {
		return deploymentState{}, fmt.Errorf("unsupported deployment state schema: %d", state.SchemaVersion)
	}
	if _, err := uuid.Parse(state.DeploymentID); err != nil {
		return deploymentState{}, fmt.Errorf("invalid deployment ID: %w", err)
	}
	if state.Mode != deploymentMode {
		return deploymentState{}, fmt.Errorf("unexpected deployment mode %q", state.Mode)
	}
	if state.CreatedAt.IsZero() {
		return deploymentState{}, errors.New("deployment creation time is missing")
	}
	return state, nil
}

func validateRuntimeEnv(path, deploymentID string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	values := parseEnv(content)
	required := []string{
		"APP_ENV",
		"LEAMOUT_DEPLOYMENT_MODE",
		"LEAMOUT_DEPLOYMENT_ID",
		"FREESWITCH_ESL_PASSWORD",
		"CARRIER_CREDENTIAL_ENCRYPTION_KEY",
		"TURN_AUTH_SECRET",
		"POSTGRES_PASSWORD",
		"REDIS_PASSWORD",
		"NATS_PASSWORD",
	}
	for _, name := range required {
		if values[name] == "" {
			return fmt.Errorf("%s is missing", name)
		}
	}
	if values["APP_ENV"] != "production" {
		return errors.New("APP_ENV must be production")
	}
	if values["LEAMOUT_DEPLOYMENT_MODE"] != deploymentMode {
		return fmt.Errorf("LEAMOUT_DEPLOYMENT_MODE must be %s", deploymentMode)
	}
	if values["LEAMOUT_DEPLOYMENT_ID"] != deploymentID {
		return errors.New("deployment ID does not match durable state")
	}
	for _, name := range []string{
		"FREESWITCH_ESL_PASSWORD",
		"TURN_AUTH_SECRET",
		"POSTGRES_PASSWORD",
		"REDIS_PASSWORD",
		"NATS_PASSWORD",
	} {
		if len(values[name]) != 64 {
			return fmt.Errorf("%s must be a 64-character generated secret", name)
		}
	}
	key, err := base64.RawURLEncoding.DecodeString(values["CARRIER_CREDENTIAL_ENCRYPTION_KEY"])
	if err != nil || len(key) != 32 {
		return errors.New("CARRIER_CREDENTIAL_ENCRYPTION_KEY must encode exactly 32 bytes")
	}
	return nil
}

func parseEnv(content []byte) map[string]string {
	values := make(map[string]string)
	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(name)] = strings.TrimSpace(value)
	}
	return values
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
