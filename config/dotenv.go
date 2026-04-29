package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

func init() {
	loadDotenv()
}

// loadDotenv runs before other packages' inits that import config. It uses
// Overload (not Load) so values from .env win over vars already set in the shell
// or in VS Code launch.json — otherwise Atlas would never replace e.g. localhost.
func loadDotenv() {
	candidates := []string{".env"}
	if _, file, _, ok := runtime.Caller(0); ok {
		// This file is config/dotenv.go → repo root is parent of config/
		repoEnv := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".env"))
		candidates = append([]string{repoEnv}, candidates...)
	}
	for _, p := range candidates {
		if err := godotenv.Overload(p); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Printf("godotenv %s: %v", p, err)
			continue
		}
		return
	}
}
