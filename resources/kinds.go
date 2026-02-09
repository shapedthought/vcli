package resources

// Kind constants for resource and mixin types
const (
	KindVBRJob                    = "VBRJob"
	KindVBRRepository             = "VBRRepository"
	KindVBRSOBR                   = "VBRSOBR"
	KindVBRScaleOutRepository     = "VBRScaleOutRepository" // Alias used in YAML specs and apply configs
	KindVBREncryptionPassword     = "VBREncryptionPassword"
	KindVBRKmsServer              = "VBRKmsServer"
	KindProfile                   = "Profile"
	KindOverlay                   = "Overlay"
)

// IsMixinKind returns true if the kind is a mixin type (Profile or Overlay)
// that is used in group merges but is not a standalone resource.
func IsMixinKind(kind string) bool {
	return kind == KindProfile || kind == KindOverlay
}

// IsResourceKind returns true if the kind represents a VBR resource type.
func IsResourceKind(kind string) bool {
	switch kind {
	case KindVBRJob, KindVBRRepository, KindVBRSOBR, KindVBRScaleOutRepository, KindVBREncryptionPassword, KindVBRKmsServer:
		return true
	default:
		return false
	}
}
