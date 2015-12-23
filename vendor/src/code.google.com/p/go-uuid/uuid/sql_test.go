package uuid

import (
	"testing"
)

func TestValuer(t *testing.T) {
	uuid := NewRandom()
	expected := uuid.String()

	result, err := uuid.Value()
	if err != nil {
		t.Error(err)
	}
	if result != expected {
		t.Error("Result should have been %v, but it was %v", expected, result)
	}
}

func TestScannerString(t *testing.T) {
	uuid := &UUID{}

	expected := New()
	err := uuid.Scan(expected)
	if err != nil {
		t.Error(err)
	}
	result := uuid.String()
	if result != expected {
		t.Errorf("Result should have been %v, but it was %v", expected, result)
	}
}

func TestScannerByteArray(t *testing.T) {
	uuid := &UUID{}

	expected := New()
	err := uuid.Scan([]byte(expected))
	if err != nil {
		t.Error(err)
	}
	result := uuid.String()
	if result != expected {
		t.Errorf("Result should have been %v, but it was %v", expected, result)
	}
}

func TestErrIncompatibleType(t *testing.T) {
	uuid := NewRandom()
	err := uuid.Scan(1)
	if err != ErrIncompatibleType {
		t.Errorf("Should have received error %v", ErrIncompatibleType)
	}
}
