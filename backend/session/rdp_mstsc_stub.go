//go:build !windows

package session

// LaunchRDP relies on the Windows mstsc.exe and is not available on this
// platform. RDP connections are handled by the embedded ActiveX path instead.
func LaunchRDP(_ ConnectionConfig) error {
	return nil
}