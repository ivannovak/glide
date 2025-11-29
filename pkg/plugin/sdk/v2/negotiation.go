package v2

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// SDKVersion is the current SDK version
const SDKVersion = "2.0.0"

// MinCompatibleSDKVersion is the minimum SDK version this implementation can work with
const MinCompatibleSDKVersion = "2.0.0"

// ProtocolVersion identifies the plugin protocol version
// This is separate from SDK version and indicates wire protocol compatibility
const ProtocolVersion = 2

// VersionInfo contains version information about a plugin and its SDK requirements
type VersionInfo struct {
	// PluginVersion is the plugin's own version
	PluginVersion string

	// SDKVersion is the SDK version the plugin was built with
	SDKVersion string

	// MinSDKVersion is the minimum SDK version the plugin requires
	// If not specified, defaults to SDKVersion
	MinSDKVersion string

	// ProtocolVersion is the protocol version the plugin uses
	ProtocolVersion int
}

// NegotiationResult contains the result of version negotiation
type NegotiationResult struct {
	// Compatible indicates if the plugin is compatible
	Compatible bool

	// Error provides details if not compatible
	Error error

	// NegotiatedProtocol is the protocol version to use
	NegotiatedProtocol int

	// RequiresAdapter indicates if an adapter layer is needed
	RequiresAdapter bool

	// AdapterType indicates which adapter to use (e.g., "v1-to-v2")
	AdapterType string
}

// NegotiateVersion determines if a plugin is compatible and how to load it
func NegotiateVersion(pluginVersion VersionInfo) *NegotiationResult {
	result := &NegotiationResult{}

	// Check protocol version
	if pluginVersion.ProtocolVersion == ProtocolVersion {
		// Native v2 plugin
		result.Compatible = true
		result.NegotiatedProtocol = ProtocolVersion
		result.RequiresAdapter = false
		return result
	}

	if pluginVersion.ProtocolVersion == 1 {
		// v1 plugin - needs adapter
		result.Compatible = true
		result.NegotiatedProtocol = 1
		result.RequiresAdapter = true
		result.AdapterType = "v1-to-v2"
		return result
	}

	// Unknown protocol version
	result.Compatible = false
	result.Error = fmt.Errorf(
		"unsupported protocol version %d (supports: 1, 2)",
		pluginVersion.ProtocolVersion,
	)
	return result
}

// CheckSDKCompatibility checks if a plugin's SDK version is compatible
func CheckSDKCompatibility(pluginSDK, hostSDK string) error {
	// Parse versions
	pluginVer, err := semver.NewVersion(pluginSDK)
	if err != nil {
		return fmt.Errorf("invalid plugin SDK version %q: %w", pluginSDK, err)
	}

	hostVer, err := semver.NewVersion(hostSDK)
	if err != nil {
		return fmt.Errorf("invalid host SDK version %q: %w", hostSDK, err)
	}

	// Check major version compatibility
	if pluginVer.Major() != hostVer.Major() {
		return fmt.Errorf(
			"SDK major version mismatch: plugin requires %d.x, host provides %d.x",
			pluginVer.Major(),
			hostVer.Major(),
		)
	}

	// Within the same major version, newer host can load older plugin
	// but older host cannot load newer plugin
	if pluginVer.GreaterThan(hostVer) {
		return fmt.Errorf(
			"plugin SDK version %s is newer than host SDK %s",
			pluginSDK,
			hostSDK,
		)
	}

	return nil
}

// GetPluginSDKVersion extracts SDK version from a plugin
// For v2 plugins, this is in metadata. For v1 plugins, we infer it.
func GetPluginSDKVersion(plugin interface{}) string {
	// Check if it's a v2 plugin
	if v2Plugin, ok := plugin.(interface{ SDKVersion() string }); ok {
		return v2Plugin.SDKVersion()
	}

	// Check if it's a V1Adapter
	if v1Adapter, ok := plugin.(*V1Adapter); ok {
		_ = v1Adapter
		// v1 plugins use SDK version 1.x
		return "1.0.0"
	}

	// Unknown plugin type, assume current SDK
	return SDKVersion
}

// VersionNegotiator handles plugin version negotiation
type VersionNegotiator struct {
	hostSDKVersion string
	hostProtocol   int
}

// NewVersionNegotiator creates a new version negotiator
func NewVersionNegotiator() *VersionNegotiator {
	return &VersionNegotiator{
		hostSDKVersion: SDKVersion,
		hostProtocol:   ProtocolVersion,
	}
}

// Negotiate performs version negotiation with a plugin
func (n *VersionNegotiator) Negotiate(info VersionInfo) *NegotiationResult {
	result := NegotiateVersion(info)

	if !result.Compatible {
		return result
	}

	// For v2 plugins, check SDK compatibility
	if result.NegotiatedProtocol == ProtocolVersion {
		minSDK := info.MinSDKVersion
		if minSDK == "" {
			minSDK = info.SDKVersion
		}

		if err := CheckSDKCompatibility(minSDK, n.hostSDKVersion); err != nil {
			result.Compatible = false
			result.Error = err
			return result
		}
	}

	return result
}

// SupportsProtocol checks if the negotiator supports a protocol version
func (n *VersionNegotiator) SupportsProtocol(version int) bool {
	// We support v1 (via adapter) and v2 (natively)
	return version == 1 || version == 2
}

// GetSupportedProtocols returns all supported protocol versions
func (n *VersionNegotiator) GetSupportedProtocols() []int {
	return []int{1, 2}
}
