package utils

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// BasicImeiCheckServiceID is imei.info "Basic IMEI Check" (slug basic-imei-check).
// Only this service may be used — premium checks (Apple Sold By, Samsung Info, etc.)
// must never be invoked from this API.
const BasicImeiCheckServiceID = "0"

// ErrorImeiInfoRateLimited is returned when paid imei.info usage exceeds configured caps.
var ErrorImeiInfoRateLimited = errors.New(
	"Se alcanzó el límite de consultas a imei.info. Intente más tarde o use un TAC ya en caché.",
)

type ImeiInfoLimits struct {
	MaxPerUserPerDay     int
	MaxGlobalPerDay      int
	CooldownSameIMEISecs int
}

func LoadImeiInfoLimits() ImeiInfoLimits {
	return ImeiInfoLimits{
		MaxPerUserPerDay:     envIntPositive("IMEI_INFO_MAX_PER_USER_DAY", 40),
		MaxGlobalPerDay:      envIntPositive("IMEI_INFO_MAX_GLOBAL_DAY", 40),
		CooldownSameIMEISecs: envIntPositive("IMEI_INFO_COOLDOWN_SAME_IMEI_SEC", 60),
	}
}

func ResolveImeiInfoServiceID() string {
	raw := strings.TrimSpace(EnvOrDefault("IMEI_INFO_SERVICE", BasicImeiCheckServiceID))
	if raw == "" {
		raw = BasicImeiCheckServiceID
	}
	if raw != BasicImeiCheckServiceID {
		// Never call premium service IDs from the server, even if misconfigured in SSM.
		return BasicImeiCheckServiceID
	}
	return raw
}

func StartOfUTCDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func envIntPositive(key string, defaultVal int) int {
	v := strings.TrimSpace(EnvOrDefault(key, ""))
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultVal
	}
	return n
}
