package infrastructure

import "time"

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	return append([]string(nil), values...)
}

func cloneFloat32s(values []float32) []float32 {
	if values == nil {
		return nil
	}
	return append([]float32(nil), values...)
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cp := *value
	return &cp
}
