package config

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

func init() {
	loadDotenv()
}

// utf8BOM is written by some Windows editors (e.g. "UTF-8 with BOM"). godotenv
// rejects it, which used to make the whole .env be silently skipped.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// loadDotenv runs before other packages' inits that import config. Values from
// .env win over vars already set in the shell or in VS Code launch.json
// (Overload semantics) — otherwise Atlas would never replace e.g. localhost.
func loadDotenv() {
	candidates := []string{".env"}
	if _, file, _, ok := runtime.Caller(0); ok {
		// This file is config/dotenv.go → repo root is parent of config/
		repoEnv := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".env"))
		candidates = append([]string{repoEnv}, candidates...)
	}
	for _, p := range candidates {
		if err := overloadEnvFile(p); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Printf("godotenv %s: %v", p, err)
			continue
		}
		return
	}
}

// overloadEnvFile parses p (tolerating a UTF-8 BOM) and sets every variable,
// overriding values already present in the environment.
func overloadEnvFile(p string) error {
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	vars, err := godotenv.UnmarshalBytes(bytes.TrimPrefix(data, utf8BOM))
	if err != nil {
		return err
	}
	for k, v := range vars {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}
