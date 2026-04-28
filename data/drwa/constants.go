package drwa

import "strings"

// DenialCode is the canonical string identifier for DRWA gate denials.
type DenialCode string

const (
	// DenialUnknown is the explicit sentinel for an unrecognized denial code.
	// The zero value remains invalid so empty or missing topics cannot be
	// mistaken for a valid DRWA gate decision.
	DenialUnknown             DenialCode = "DRWA_UNKNOWN"
	DenialPolicyNotSynced     DenialCode = "DRWA_POLICY_NOT_SYNCED"
	DenialTokenPaused         DenialCode = "DRWA_TOKEN_PAUSED"
	DenialKYCRequiredSender   DenialCode = "DRWA_KYC_REQUIRED_SENDER"
	DenialAMLBlockedSender    DenialCode = "DRWA_AML_BLOCKED_SENDER"
	DenialAssetExpired        DenialCode = "DRWA_ASSET_EXPIRED"
	DenialTransferLocked      DenialCode = "DRWA_TRANSFER_LOCKED"
	DenialKYCRequiredReceiver DenialCode = "DRWA_KYC_REQUIRED_RECEIVER"
	DenialAMLBlockedReceiver  DenialCode = "DRWA_AML_BLOCKED_RECEIVER"
	DenialReceiveLocked       DenialCode = "DRWA_RECEIVE_LOCKED"
	DenialInvestorClass       DenialCode = "DRWA_INVESTOR_CLASS_BLOCKED"
	DenialJurisdiction        DenialCode = "DRWA_JURISDICTION_BLOCKED"
	DenialAuditorRequired     DenialCode = "DRWA_AUDITOR_REQUIRED"
	DenialTravelRuleRequired  DenialCode = "DRWA_TRAVEL_RULE_REQUIRED"
	DenialSanctionsMatch      DenialCode = "DRWA_SANCTIONS_MATCH"
	DenialWindDownActive      DenialCode = "DRWA_WIND_DOWN_ACTIVE"
)

// AllDenialCodes returns the canonical concrete DRWA denial codes in stable
// order. The DenialUnknown sentinel is intentionally excluded so consumers can
// use this as an exhaustiveness contract for emitted gate decisions.
func AllDenialCodes() []DenialCode {
	return []DenialCode{
		DenialPolicyNotSynced,
		DenialTokenPaused,
		DenialKYCRequiredSender,
		DenialAMLBlockedSender,
		DenialAssetExpired,
		DenialTransferLocked,
		DenialKYCRequiredReceiver,
		DenialAMLBlockedReceiver,
		DenialReceiveLocked,
		DenialInvestorClass,
		DenialJurisdiction,
		DenialAuditorRequired,
		DenialTravelRuleRequired,
		DenialSanctionsMatch,
		DenialWindDownActive,
	}
}

// IsKnown reports whether code is one of the concrete denial codes emitted by
// the DRWA gate.
func (code DenialCode) IsKnown() bool {
	for _, known := range AllDenialCodes() {
		if code == known {
			return true
		}
	}
	return false
}

// IsValid reports whether code is one of the concrete canonical denial values.
// The explicit unknown sentinel is a normalization result, not a storable
// denial reason.
func (code DenialCode) IsValid() bool {
	return code.IsKnown()
}

// NormalizeDenialCode canonicalizes a raw denial code string. Unknown non-empty
// values map to DenialUnknown instead of leaking arbitrary strings into
// downstream typed surfaces.
func NormalizeDenialCode(raw string) DenialCode {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	code := DenialCode(trimmed)
	if code.IsKnown() {
		return code
	}

	upper := DenialCode(strings.ToUpper(trimmed))
	if upper.IsKnown() {
		return upper
	}

	return DenialUnknown
}

// StorageKeyPrefix is the typed canonical prefix for DRWA account-data keys.
type StorageKeyPrefix string

// String returns the raw prefix for storage-key construction.
func (prefix StorageKeyPrefix) String() string {
	return string(prefix)
}

// Storage key prefixes used by DRWA in the account data trie.
const (
	TokenPolicyPrefix       StorageKeyPrefix = "drwa:token:"
	HolderMirrorPrefix      StorageKeyPrefix = "drwa:holder:"
	HolderProfilePrefix     StorageKeyPrefix = "drwa:profile:"
	HolderAuditorAuthPrefix StorageKeyPrefix = "drwa:auditor:"
	AssetRecordPrefix       StorageKeyPrefix = "drwa:asset:"
)

// AllStorageKeyPrefixes returns the complete DRWA storage prefix set in stable
// order so cross-repo consumers can test exhaustiveness.
func AllStorageKeyPrefixes() []StorageKeyPrefix {
	return []StorageKeyPrefix{
		TokenPolicyPrefix,
		HolderMirrorPrefix,
		HolderProfilePrefix,
		HolderAuditorAuthPrefix,
		AssetRecordPrefix,
	}
}

// IsValid reports whether prefix is one of the canonical DRWA storage prefixes.
func (prefix StorageKeyPrefix) IsValid() bool {
	for _, known := range AllStorageKeyPrefixes() {
		if prefix == known {
			return true
		}
	}
	return false
}
