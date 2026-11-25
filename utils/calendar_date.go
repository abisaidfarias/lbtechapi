package utils

import "time"

// NormalizeCalendarDateUTC stores calendar dates at noon UTC so clients in different
// timezones show the same day when rendering date-only fields.
func NormalizeCalendarDateUTC(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 12, 0, 0, 0, time.UTC)
}

func NormalizeOptionalCalendarDateUTC(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	normalized := NormalizeCalendarDateUTC(*value)
	return &normalized
}

// OptionalTimeForBSON stores nil or a zero time as null and any other value
// unchanged (no calendar normalization).
func OptionalTimeForBSON(value *time.Time) interface{} {
	if value == nil || value.IsZero() {
		return nil
	}
	return *value
}

// TimePtrOrNil returns nil for a zero time, a pointer to it otherwise.
func TimePtrOrNil(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func OptionalDateForBSON(value *time.Time) interface{} {
	if value == nil || value.IsZero() {
		return nil
	}
	return NormalizeCalendarDateUTC(*value)
}
