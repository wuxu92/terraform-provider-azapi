package nativeacc

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func TestDumpTFConfigEnabled(t *testing.T) {
	t.Setenv(dumpTFConfigEnv, "")
	if dumpTFConfigEnabled() {
		t.Fatal("empty env should disable config dump")
	}

	t.Setenv(dumpTFConfigEnv, "true")
	if !dumpTFConfigEnabled() {
		t.Fatal("true env should enable config dump")
	}

	t.Setenv(dumpTFConfigEnv, "1")
	if !dumpTFConfigEnabled() {
		t.Fatal("1 env should enable config dump")
	}

	t.Setenv(dumpTFConfigEnv, "false")
	if dumpTFConfigEnabled() {
		t.Fatal("false env should disable config dump")
	}

	t.Setenv(dumpTFConfigOnlyEnv, "true")
	t.Setenv(dumpTFConfigEnv, "")
	if !dumpTFConfigOnlyEnabled() {
		t.Fatal("dump-only env should enable dump-only mode")
	}
	if !dumpTFConfigEnabled() {
		t.Fatal("dump-only env should imply config dump")
	}

	t.Setenv(dumpTFConfigOnlyEnv, "false")
	if dumpTFConfigOnlyEnabled() {
		t.Fatal("false dump-only env should disable dump-only mode")
	}
}

func TestTFConfigDumpIncludesWholeWorkspaceConfig(t *testing.T) {
	dir := t.TempDir()
	mustWrite := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	mustWrite("provider.tf", "provider \"azapi\" {}\n")
	mustWrite("resource_z.tf", "resource \"azapi_resource_group\" \"rg\" {}\n")
	mustWrite("terraform.tfstate", "not config")

	dump, err := (&Workspace{dir: dir}).tfConfigDump("unit test")
	if err != nil {
		t.Fatalf("tfConfigDump: %v", err)
	}

	for _, want := range []string{
		"===== BEGIN Terraform config dump: unit test =====",
		"# Workspace: " + dir,
		"----- provider.tf -----\nprovider \"azapi\" {}\n",
		"----- resource_z.tf -----\nresource \"azapi_resource_group\" \"rg\" {}\n",
		"===== END Terraform config dump: unit test =====",
	} {
		if !strings.Contains(dump, want) {
			t.Fatalf("dump missing %q in:\n%s", want, dump)
		}
	}
	if strings.Contains(dump, "terraform.tfstate") || strings.Contains(dump, "not config") {
		t.Fatalf("dump should include only .tf config files:\n%s", dump)
	}
	if strings.Index(dump, "provider.tf") > strings.Index(dump, "resource_z.tf") {
		t.Fatalf("dump files should be sorted:\n%s", dump)
	}
}

func TestDumpTFConfigIfEnabledAppendsConfiguredFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "provider.tf"), []byte("provider \"azapi\" {}\n"), 0o600); err != nil {
		t.Fatalf("writing provider.tf: %v", err)
	}
	out := filepath.Join(t.TempDir(), "nested", "workspace.tf.dump")
	t.Setenv(dumpTFConfigEnv, "true")
	t.Setenv(dumpTFConfigFileEnv, out)

	(&Workspace{dir: dir}).dumpTFConfigIfEnabled("before terraform apply (unit)")
	(&Workspace{dir: dir}).dumpTFConfigIfEnabled("before terraform apply (unit second)")

	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading dump file: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"===== BEGIN Terraform config dump: before terraform apply (unit) =====",
		"# Workspace: " + dir,
		"----- provider.tf -----\nprovider \"azapi\" {}\n",
		"===== END Terraform config dump: before terraform apply (unit) =====",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dump file missing %q in:\n%s", want, text)
		}
	}
	if got := strings.Count(text, "===== BEGIN Terraform config dump:"); got != 2 {
		t.Fatalf("dump file contains %d dumps, want 2:\n%s", got, text)
	}
	if !strings.Contains(text, "===== BEGIN Terraform config dump: before terraform apply (unit second) =====") {
		t.Fatalf("dump file missing second appended dump:\n%s", text)
	}
}
