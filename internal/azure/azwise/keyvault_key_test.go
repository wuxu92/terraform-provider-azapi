package azwise

import (
	"testing"
	"time"
)

func TestKeyVaultKeyImplementsInterface(t *testing.T) {
	var _ ResourceKnowledge = (*KeyVaultKey)(nil)
	k := NewKeyVaultKey()
	if k.GetResourceType() != "Microsoft.KeyVault/vaults/keys" {
		t.Errorf("unexpected resource type: %s", k.GetResourceType())
	}
}

func TestKeyVaultKeyForceNew(t *testing.T) {
	k := NewKeyVaultKey()

	tests := []struct {
		name     string
		old, new map[string]interface{}
		want     bool
	}{
		{
			name: "name changed",
			old:  map[string]interface{}{"name": "key-a"},
			new:  map[string]interface{}{"name": "key-b"},
			want: true,
		},
		{
			name: "key type changed",
			old:  map[string]interface{}{"properties": map[string]interface{}{"kty": "RSA"}},
			new:  map[string]interface{}{"properties": map[string]interface{}{"kty": "EC"}},
			want: true,
		},
		{
			name: "key size changed",
			old:  map[string]interface{}{"properties": map[string]interface{}{"keySize": float64(2048)}},
			new:  map[string]interface{}{"properties": map[string]interface{}{"keySize": float64(4096)}},
			want: true,
		},
		{
			name: "curve changed",
			old:  map[string]interface{}{"properties": map[string]interface{}{"curveName": "P-256"}},
			new:  map[string]interface{}{"properties": map[string]interface{}{"curveName": "P-384"}},
			want: true,
		},
		{
			name: "no change",
			old:  map[string]interface{}{"name": "key-a", "properties": map[string]interface{}{"kty": "RSA"}},
			new:  map[string]interface{}{"name": "key-a", "properties": map[string]interface{}{"kty": "RSA"}},
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

func TestKeyVaultKeyValidateName(t *testing.T) {
	k := NewKeyVaultKey()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid alphanumeric", "my-key-01", false},
		{"valid single char", "a", false},
		{"max length 127", string(make([]byte, 127)), true}, // all null bytes won't match regex
		{"valid 127 chars", func() string {
			b := make([]byte, 127)
			for i := range b {
				b[i] = 'a'
			}
			return string(b)
		}(), false},
		{"too long 128", func() string {
			b := make([]byte, 128)
			for i := range b {
				b[i] = 'a'
			}
			return string(b)
		}(), true},
		{"empty", "", true},
		{"underscore rejected", "my_key", true},
		{"space rejected", "my key", true},
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

func TestKeyVaultKeyValidateProperties(t *testing.T) {
	k := NewKeyVaultKey()

	tests := []struct {
		name      string
		body      map[string]interface{}
		wantCount int
	}{
		{
			name: "valid RSA key",
			body: map[string]interface{}{
				"properties": map[string]interface{}{
					"kty":    "RSA",
					"keyOps": []interface{}{"encrypt", "decrypt"},
				},
			},
			wantCount: 0,
		},
		{
			name: "valid EC key with curve",
			body: map[string]interface{}{
				"properties": map[string]interface{}{
					"kty":       "EC",
					"curveName": "P-256",
					"keyOps":    []interface{}{"sign", "verify"},
				},
			},
			wantCount: 0,
		},
		{
			name: "valid management-plane import and release operations",
			body: map[string]interface{}{
				"properties": map[string]interface{}{
					"kty":    "RSA",
					"keyOps": []interface{}{"import", "release"},
				},
			},
			wantCount: 0,
		},
		{
			name: "invalid key type",
			body: map[string]interface{}{
				"properties": map[string]interface{}{
					"kty": "AES",
				},
			},
			wantCount: 1,
		},
		{
			name: "invalid curve",
			body: map[string]interface{}{
				"properties": map[string]interface{}{
					"curveName": "P-192",
				},
			},
			wantCount: 1,
		},
		{
			name: "rejects non-ARM key operation",
			body: map[string]interface{}{
				"properties": map[string]interface{}{
					"keyOps": []interface{}{"encrypt", "backup"},
				},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := k.ValidateProperties(tt.body)
			if len(errs) != tt.wantCount {
				t.Errorf("ValidateProperties() got %d errors, want %d: %v", len(errs), tt.wantCount, errs)
			}
		})
	}
}

func TestKeyVaultKeyTimeouts(t *testing.T) {
	k := NewKeyVaultKey()

	for _, op := range []string{"create", "read", "update", "delete"} {
		t.Run(op, func(t *testing.T) {
			got := k.TimeoutDefault(op, 5*time.Minute)
			if got != 30*time.Minute {
				t.Errorf("TimeoutDefault(%s) = %v, want 30m", op, got)
			}
		})
	}
}

func TestKeyVaultKeySoftDelete(t *testing.T) {
	k := NewKeyVaultKey()
	if !k.IsSoftDelete() {
		t.Error("expected soft delete to be true for keyvault keys")
	}
}
