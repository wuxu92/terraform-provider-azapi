package nativeacc

import (
	"io"
	"log"
	"os"
	"testing"
)

// TestQuietProviderLogs verifies the in-process provider's log streams are silenced
// by default, that TF_LOG opts back in, and that an explicit per-stream level wins.
func TestQuietProviderLogs(t *testing.T) {
	clear := func(t *testing.T) {
		// t.Setenv("") makes os.Getenv return "" — the "unset" the helper keys off —
		// and restores the prior value when the subtest ends.
		for _, v := range append([]string{"TF_LOG"}, providerLogEnvVars...) {
			t.Setenv(v, "")
		}
	}

	t.Run("default silences every stream", func(t *testing.T) {
		clear(t)
		quietProviderLogs()
		for _, v := range providerLogEnvVars {
			if got := os.Getenv(v); got != "off" {
				t.Errorf("%s = %q, want off", v, got)
			}
		}
	})

	t.Run("TF_LOG opts back in", func(t *testing.T) {
		clear(t)
		t.Setenv("TF_LOG", "debug")
		quietProviderLogs()
		for _, v := range providerLogEnvVars {
			if got := os.Getenv(v); got != "debug" {
				t.Errorf("%s = %q, want debug", v, got)
			}
		}
	})

	t.Run("explicit per-stream level wins", func(t *testing.T) {
		clear(t)
		t.Setenv("TF_LOG_SDK_PROTO", "trace")
		quietProviderLogs()
		if got := os.Getenv("TF_LOG_SDK_PROTO"); got != "trace" {
			t.Errorf("TF_LOG_SDK_PROTO = %q, want trace (preserved)", got)
		}
		if got := os.Getenv("TF_LOG_PROVIDER"); got != "off" {
			t.Errorf("TF_LOG_PROVIDER = %q, want off", got)
		}
	})

	t.Run("root provider stream gated by name-suffixed var", func(t *testing.T) {
		// @module=provider (tflog) is gated by ToUpper("TF_LOG_PROVIDER_"+addr); DebugServe
		// defaults addr to "provider", so the real gate is TF_LOG_PROVIDER_PROVIDER — bare
		// TF_LOG_PROVIDER does NOT silence the in-process provider stream.
		const gate = "TF_LOG_PROVIDER_PROVIDER"
		found := false
		for _, v := range providerLogEnvVars {
			if v == gate {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s missing from providerLogEnvVars %v", gate, providerLogEnvVars)
		}
		clear(t)
		quietProviderLogs()
		if got := os.Getenv(gate); got != "off" {
			t.Errorf("%s = %q, want off", gate, got)
		}
	})

	t.Run("stdlib log (azure HTTP) discarded by default, TF_LOG restores", func(t *testing.T) {
		orig := log.Writer()
		t.Cleanup(func() { log.SetOutput(orig) })

		clear(t)
		quietProviderLogs()
		if log.Writer() != io.Discard {
			t.Errorf("stdlib log writer = %v, want io.Discard", log.Writer())
		}

		t.Setenv("TF_LOG", "debug")
		quietProviderLogs()
		if log.Writer() != os.Stderr {
			t.Errorf("stdlib log writer = %v, want os.Stderr", log.Writer())
		}
	})
}
