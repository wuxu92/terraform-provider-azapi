package azwise

import (
	"testing"
	"time"
)

func TestKeyVaultSecretImplementsInterface(t *testing.T) {
	var _ ResourceKnowledge = (*KeyVaultSecret)(nil)
	k := NewKeyVaultSecret()
	if k.GetResourceType() != "Microsoft.KeyVault/vaults/secrets" {
		t.Errorf("unexpected resource type: %s", k.GetResourceType())
	}
}

func TestKeyVaultSecretForceNew(t *testing.T) {
	k := NewKeyVaultSecret()

	tests := []struct {
		name     string
		old, new map[string]interface{}
		want     bool
	}{
		{
			name: "name changed",
			old:  map[string]interface{}{"name": "secret-a"},
			new:  map[string]interface{}{"name": "secret-b"},
			want: true,
		},
		{
			name: "value changed is not force new",
			old:  map[string]interface{}{"properties": map[string]interface{}{"value": "old"}},
			new:  map[string]interface{}{"properties": map[string]interface{}{"value": "new"}},
			want: false,
		},
		{
			name: "no change",
			old:  map[string]interface{}{"name": "secret-a"},
			new:  map[string]interface{}{"name": "secret-a"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := k.CheckForceNew(tt.old, tt.new); got != tt.want {
				t.Errorf("CheckForceNew() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyVaultSecretValidateName(t *testing.T) {
	k := NewKeyVaultSecret()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "my-secret-01", false},
		{"underscore rejected", "my_secret", true},
		{"empty rejected", "", true},
		{"too long 128", func() string {
			b := make([]byte, 128)
			for i := range b {
				b[i] = 'a'
			}
			return string(b)
		}(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestKeyVaultSecretSensitiveFields(t *testing.T) {
	k := NewKeyVaultSecret()

	diags := k.Validate("my-secret", map[string]interface{}{
		"properties": map[string]interface{}{"value": "hunter2"},
	}, false)

	// Without sensitive body flag, should warn about sensitive field usage
	if !diags.HasError() {
		t.Error("expected diagnostic about sensitive fields when hasSensitiveBody is false")
	}

	// With sensitive body flag, no diagnostic
	diags = k.Validate("my-secret", map[string]interface{}{
		"properties": map[string]interface{}{"value": "hunter2"},
	}, true)

	if diags.HasError() {
		for _, d := range diags {
			t.Errorf("unexpected diagnostic with hasSensitiveBody=true: %s: %s", d.Summary(), d.Detail())
		}
	}
}

func TestKeyVaultSecretTimeouts(t *testing.T) {
	k := NewKeyVaultSecret()

	for _, op := range []string{"create", "read", "update", "delete"} {
		t.Run(op, func(t *testing.T) {
			got := k.TimeoutDefault(op, 5*time.Minute)
			if got != 30*time.Minute {
				t.Errorf("TimeoutDefault(%s) = %v, want 30m", op, got)
			}
		})
	}
}

func TestKeyVaultSecretSoftDelete(t *testing.T) {
	k := NewKeyVaultSecret()
	if !k.IsSoftDelete() {
		t.Error("expected soft delete to be true for keyvault secrets")
	}
}
