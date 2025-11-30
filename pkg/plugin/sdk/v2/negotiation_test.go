package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNegotiateVersion_V2Plugin(t *testing.T) {
	info := VersionInfo{
		PluginVersion:   "1.0.0",
		SDKVersion:      "2.0.0",
		ProtocolVersion: 2,
	}

	result := NegotiateVersion(info)

	assert.True(t, result.Compatible)
	assert.NoError(t, result.Error)
	assert.Equal(t, 2, result.NegotiatedProtocol)
	assert.False(t, result.RequiresAdapter)
	assert.Empty(t, result.AdapterType)
}

func TestNegotiateVersion_V1Plugin(t *testing.T) {
	info := VersionInfo{
		PluginVersion:   "1.0.0",
		SDKVersion:      "1.0.0",
		ProtocolVersion: 1,
	}

	result := NegotiateVersion(info)

	assert.True(t, result.Compatible)
	assert.NoError(t, result.Error)
	assert.Equal(t, 1, result.NegotiatedProtocol)
	assert.True(t, result.RequiresAdapter)
	assert.Equal(t, "v1-to-v2", result.AdapterType)
}

func TestNegotiateVersion_UnsupportedProtocol(t *testing.T) {
	info := VersionInfo{
		PluginVersion:   "3.0.0",
		SDKVersion:      "3.0.0",
		ProtocolVersion: 3,
	}

	result := NegotiateVersion(info)

	assert.False(t, result.Compatible)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "unsupported protocol version")
}

func TestCheckSDKCompatibility_SameVersion(t *testing.T) {
	err := CheckSDKCompatibility("2.0.0", "2.0.0")
	assert.NoError(t, err)
}

func TestCheckSDKCompatibility_HostNewer(t *testing.T) {
	// Host v2.1.0 can load plugin built with v2.0.0
	err := CheckSDKCompatibility("2.0.0", "2.1.0")
	assert.NoError(t, err)
}

func TestCheckSDKCompatibility_PluginNewer(t *testing.T) {
	// Host v2.0.0 cannot load plugin built with v2.1.0
	err := CheckSDKCompatibility("2.1.0", "2.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin SDK version")
	assert.Contains(t, err.Error(), "newer than host")
}

func TestCheckSDKCompatibility_DifferentMajor(t *testing.T) {
	// Major version mismatch
	err := CheckSDKCompatibility("1.0.0", "2.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "major version mismatch")
}

func TestCheckSDKCompatibility_InvalidPluginVersion(t *testing.T) {
	err := CheckSDKCompatibility("invalid", "2.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid plugin SDK version")
}

func TestCheckSDKCompatibility_InvalidHostVersion(t *testing.T) {
	err := CheckSDKCompatibility("2.0.0", "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid host SDK version")
}

func TestGetPluginSDKVersion_V2Plugin(t *testing.T) {
	plugin := NewTestPlugin()

	// TestPlugin doesn't implement SDKVersion(), so it should return current SDK
	version := GetPluginSDKVersion(plugin)
	assert.Equal(t, SDKVersion, version)
}

func TestGetPluginSDKVersion_V1Adapter(t *testing.T) {
	v1Plugin := &MockV1InProcessPlugin{name: "test"}
	adapter := AdaptV1InProcessPlugin(v1Plugin)

	version := GetPluginSDKVersion(adapter)
	assert.Equal(t, "1.0.0", version)
}

func TestVersionNegotiator_Negotiate_V2Plugin(t *testing.T) {
	negotiator := NewVersionNegotiator()

	info := VersionInfo{
		PluginVersion:   "1.5.0",
		SDKVersion:      "2.0.0",
		MinSDKVersion:   "2.0.0",
		ProtocolVersion: 2,
	}

	result := negotiator.Negotiate(info)

	assert.True(t, result.Compatible)
	assert.NoError(t, result.Error)
	assert.Equal(t, 2, result.NegotiatedProtocol)
	assert.False(t, result.RequiresAdapter)
}

func TestVersionNegotiator_Negotiate_V2Plugin_MinSDKDefault(t *testing.T) {
	negotiator := NewVersionNegotiator()

	info := VersionInfo{
		PluginVersion:   "1.5.0",
		SDKVersion:      "2.0.0",
		MinSDKVersion:   "", // Not specified, should default to SDKVersion
		ProtocolVersion: 2,
	}

	result := negotiator.Negotiate(info)

	assert.True(t, result.Compatible)
	assert.NoError(t, result.Error)
}

func TestVersionNegotiator_Negotiate_V2Plugin_IncompatibleSDK(t *testing.T) {
	negotiator := NewVersionNegotiator()

	info := VersionInfo{
		PluginVersion:   "1.0.0",
		SDKVersion:      "3.0.0", // Future SDK version
		MinSDKVersion:   "3.0.0",
		ProtocolVersion: 2,
	}

	result := negotiator.Negotiate(info)

	assert.False(t, result.Compatible)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "major version mismatch")
}

func TestVersionNegotiator_Negotiate_V1Plugin(t *testing.T) {
	negotiator := NewVersionNegotiator()

	info := VersionInfo{
		PluginVersion:   "1.0.0",
		SDKVersion:      "1.0.0",
		ProtocolVersion: 1,
	}

	result := negotiator.Negotiate(info)

	assert.True(t, result.Compatible)
	assert.NoError(t, result.Error)
	assert.Equal(t, 1, result.NegotiatedProtocol)
	assert.True(t, result.RequiresAdapter)
	assert.Equal(t, "v1-to-v2", result.AdapterType)
}

func TestVersionNegotiator_Negotiate_UnsupportedProtocol(t *testing.T) {
	negotiator := NewVersionNegotiator()

	info := VersionInfo{
		PluginVersion:   "99.0.0",
		SDKVersion:      "99.0.0",
		ProtocolVersion: 99,
	}

	result := negotiator.Negotiate(info)

	assert.False(t, result.Compatible)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "unsupported protocol version")
}

func TestVersionNegotiator_SupportsProtocol(t *testing.T) {
	negotiator := NewVersionNegotiator()

	assert.True(t, negotiator.SupportsProtocol(1))
	assert.True(t, negotiator.SupportsProtocol(2))
	assert.False(t, negotiator.SupportsProtocol(0))
	assert.False(t, negotiator.SupportsProtocol(3))
}

func TestVersionNegotiator_GetSupportedProtocols(t *testing.T) {
	negotiator := NewVersionNegotiator()

	protocols := negotiator.GetSupportedProtocols()
	require.Len(t, protocols, 2)
	assert.Contains(t, protocols, 1)
	assert.Contains(t, protocols, 2)
}

func TestSDKVersionConstants(t *testing.T) {
	// Verify constants are defined and valid semver
	assert.NotEmpty(t, SDKVersion)
	assert.NotEmpty(t, MinCompatibleSDKVersion)
	assert.Greater(t, ProtocolVersion, 0)

	// Verify SDKVersion is valid semver
	err := CheckSDKCompatibility(SDKVersion, SDKVersion)
	assert.NoError(t, err)

	// Verify MinCompatibleSDKVersion is valid semver
	err = CheckSDKCompatibility(MinCompatibleSDKVersion, MinCompatibleSDKVersion)
	assert.NoError(t, err)
}

func TestVersionInfo_Complete(t *testing.T) {
	info := VersionInfo{
		PluginVersion:   "1.2.3",
		SDKVersion:      "2.0.0",
		MinSDKVersion:   "2.0.0",
		ProtocolVersion: 2,
	}

	assert.Equal(t, "1.2.3", info.PluginVersion)
	assert.Equal(t, "2.0.0", info.SDKVersion)
	assert.Equal(t, "2.0.0", info.MinSDKVersion)
	assert.Equal(t, 2, info.ProtocolVersion)
}

func TestNegotiationResult_Complete(t *testing.T) {
	result := &NegotiationResult{
		Compatible:         true,
		Error:              nil,
		NegotiatedProtocol: 2,
		RequiresAdapter:    false,
		AdapterType:        "",
	}

	assert.True(t, result.Compatible)
	assert.NoError(t, result.Error)
	assert.Equal(t, 2, result.NegotiatedProtocol)
	assert.False(t, result.RequiresAdapter)
	assert.Empty(t, result.AdapterType)
}
