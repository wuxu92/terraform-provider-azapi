package azwise

import "testing"

func TestBlobServiceImplementsInterface(t *testing.T) {
	var _ ResourceKnowledge = NewStorageAccountBlobService()
}

// TestBlobServiceValidateProperties exercises the declarative rules ported from
// AzureRM's blob_properties block: retention-day IntBetween ranges, the
// defaultServiceVersion enum, and the cors MaxItems cap.
func TestBlobServiceValidateProperties(t *testing.T) {
	k := NewStorageAccountBlobService()

	tests := []struct {
		name      string
		body      map[string]interface{}
		wantCount int
	}{
		{
			"valid retention and version",
			map[string]interface{}{
				"properties": map[string]interface{}{
					"deleteRetentionPolicy": map[string]interface{}{"days": 30},
					"defaultServiceVersion": "2023-01-03",
				},
			},
			0,
		},
		{
			"delete retention days too high",
			map[string]interface{}{
				"properties": map[string]interface{}{
					"deleteRetentionPolicy": map[string]interface{}{"days": 400},
				},
			},
			1,
		},
		{
			"change feed retention out of range",
			map[string]interface{}{
				"properties": map[string]interface{}{
					"changeFeed": map[string]interface{}{"retentionInDays": 200000},
				},
			},
			1,
		},
		{
			"invalid default service version",
			map[string]interface{}{
				"properties": map[string]interface{}{
					"defaultServiceVersion": "not-a-version",
				},
			},
			1,
		},
		{
			"too many cors rules",
			map[string]interface{}{
				"properties": map[string]interface{}{
					"cors": map[string]interface{}{
						"corsRules": []interface{}{1, 2, 3, 4, 5, 6},
					},
				},
			},
			1,
		},
		{
			"absent properties",
			map[string]interface{}{},
			0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := k.ValidateProperties(tc.body)
			if len(errs) != tc.wantCount {
				t.Errorf("ValidateProperties() got %d errors, want %d: %v", len(errs), tc.wantCount, errs)
			}
		})
	}
}
