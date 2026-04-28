package drwa

import "testing"

func TestDenialCodes_NonEmpty(t *testing.T) {
	for _, code := range AllDenialCodes() {
		if code == "" {
			t.Fatalf("denial code must not be empty")
		}
	}
}

func TestDenialCodes_Unique(t *testing.T) {
	codes := AllDenialCodes()
	seen := make(map[DenialCode]struct{}, len(codes))
	for _, code := range codes {
		if _, exists := seen[code]; exists {
			t.Fatalf("duplicate denial code %q", code)
		}
		seen[code] = struct{}{}
	}
}

func TestDenialCodes_ValidityAndNormalization(t *testing.T) {
	if DenialCode("").IsValid() {
		t.Fatalf("empty denial code must be invalid")
	}
	if DenialUnknown.IsValid() {
		t.Fatalf("unknown sentinel must not be valid for storage")
	}
	if DenialUnknown.IsKnown() {
		t.Fatalf("unknown sentinel must not be a concrete known denial")
	}
	for _, code := range AllDenialCodes() {
		if !code.IsKnown() {
			t.Fatalf("denial code %q must be known", code)
		}
		if !code.IsValid() {
			t.Fatalf("denial code %q must be valid", code)
		}
	}

	if got := NormalizeDenialCode("  drwa_kyc_required_sender  "); got != DenialKYCRequiredSender {
		t.Fatalf("expected canonical upper-case sender KYC code, got %q", got)
	}
	if got := NormalizeDenialCode("  DRWA_NOT_REAL  "); got != DenialUnknown {
		t.Fatalf("expected unknown sentinel, got %q", got)
	}
	if got := NormalizeDenialCode("  "); got != "" {
		t.Fatalf("expected empty normalization for blank input, got %q", got)
	}
}

func TestPrefixes_NonEmpty(t *testing.T) {
	prefixes := AllStorageKeyPrefixes()
	for _, p := range prefixes {
		if p == "" {
			t.Fatalf("prefix must not be empty")
		}
	}
}

func TestPrefixes_UniqueAndNonOverlapping(t *testing.T) {
	prefixes := AllStorageKeyPrefixes()
	seen := make(map[StorageKeyPrefix]struct{}, len(prefixes))
	for _, prefix := range prefixes {
		if !prefix.IsValid() {
			t.Fatalf("prefix %q must be valid", prefix)
		}
		if _, exists := seen[prefix]; exists {
			t.Fatalf("duplicate prefix %q", prefix)
		}
		seen[prefix] = struct{}{}
	}
	for i, left := range prefixes {
		for j, right := range prefixes {
			if i == j {
				continue
			}
			leftRaw := left.String()
			rightRaw := right.String()
			if len(leftRaw) <= len(rightRaw) && leftRaw == rightRaw[:len(leftRaw)] {
				t.Fatalf("prefix %q overlaps with %q", left, right)
			}
		}
	}
}

func TestPrefixes_RejectUnknown(t *testing.T) {
	if StorageKeyPrefix("").IsValid() {
		t.Fatalf("empty prefix must be invalid")
	}
	if StorageKeyPrefix("drwa:unknown:").IsValid() {
		t.Fatalf("unknown prefix must be invalid")
	}
}
