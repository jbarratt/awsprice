package awsprice

import "testing"

func TestRegion(t *testing.T) {
	expected := Region("US West (N. California)")
	got, err := NewRegion("us-west-1")

	if err != nil {
		t.Error("Error creating us-west-1")
	}

	if expected != got {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}
