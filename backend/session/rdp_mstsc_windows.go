//go:build windows

package session

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/ys-ll/uniterm/backend/log"
	"golang.org/x/sys/windows"
)

// RDP auto-logon via the system mstsc.exe.
//
// The embedded MS RDP ActiveX control cannot auto-logon silently under
// NLA/CredSSP — it ignores both IMsTscNonScriptable.ClearTextPassword and a
// pre-written "TERMSRV/<host>" Credential Manager entry, and still shows the
// Windows credential prompt. mstsc itself auto-logs on because it drives its
// own credential pipeline (reading the Credential Manager). So we hand the
// connection to mstsc with a minimal .rdp profile and a pre-written
// "TERMSRV/<host>" vault credential.

func splitRdpDomainUser(user string) (domain, name string) {
	if i := strings.LastIndex(user, `\`); i >= 0 {
		return user[:i], user[i+1:]
	}
	return "", user
}

func buildRDPFile(c ConnectionConfig) string {
	port := c.Port
	if port <= 0 {
		port = 3389
	}
	domain, user := c.RdpDomain, c.User
	if domain == "" {
		domain, user = splitRdpDomainUser(c.User)
	}

	// Deliberately password-free: a `password 51:b:` line makes mstsc error on
	// load ("从文件…rdp 加载时出错") on some Windows builds, so auto-logon under
	// NLA is handled by the pre-written "TERMSRV/<host>" Credential Manager
	// entry, which mstsc natively reads — the same mechanism its own "Remember
	// credentials" uses. The lines below are standard mstsc .rdp settings.
	lines := []string{
		fmt.Sprintf("full address:s:%s", c.Host),
		fmt.Sprintf("port:i:%d", port),
		fmt.Sprintf("username:s:%s", user),
		// NLA on (matches the legacy "enablecredsspsupport" + auth level).
		"enablecredsspsupport:i:1",
		"authentication level:i:2",
	}

	// Display settings from the connection form. -1 (or unset) means fullscreen,
	// mirroring rdpFixedWidth==-1 in the embedded path; a positive value is a
	// fixed windowed resolution.
	if c.RdpFixedWidth == -1 || c.RdpFixedHeight == -1 {
		lines = append(lines, "screen mode id:i:2") // full screen
	} else {
		w, h := c.RdpFixedWidth, c.RdpFixedHeight
		if w <= 0 {
			w = 1280
		}
		if h <= 0 {
			h = 800
		}
		lines = append(lines,
			"screen mode id:i:1", // window
			fmt.Sprintf("desktopwidth:i:%d", w),
			fmt.Sprintf("desktopheight:i:%d", h),
		)
	}
	if c.RdpSmartSizing {
		lines = append(lines, "smart sizing:i:1")
	}

	if domain != "" {
		lines = append(lines, "domain:s:"+domain)
	}
	return strings.Join(lines, "\r\n")
}

// LaunchRDP writes a temporary .rdp profile for the connection and hands it to
// the system mstsc.exe so it auto-logs on (honoring NLA and the saved
// password). The temporary file is removed shortly after mstsc has read it.
func LaunchRDP(config ConnectionConfig) error {
	content := buildRDPFile(config)

	f, err := os.CreateTemp("", "uniterm-*.rdp")
	if err != nil {
		return fmt.Errorf("create temp rdp: %w", err)
	}
	path := f.Name()
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(path)
		return fmt.Errorf("write temp rdp: %w", err)
	}
	f.Close()

	// Ensure mstsc has the network-level logon credential: write (or refresh)
	// the "TERMSRV/<host>" Credential Manager entry so CredSSP auto-logs in
	// under NLA, exactly as mstsc's own "Remember credentials" does.
	domain, user := config.RdpDomain, config.User
	if domain == "" {
		domain, user = splitRdpDomainUser(config.User)
	}
	rdpSyncCredential(config.Host, domain, user, config.Password)

	cmd := "mstsc.exe"

	// ShellExecuteW ("open" + SW_SHOWNORMAL) launches mstsc with the profile and
	// activates its window, so the RDP window surfaces in front of uniTerm — a
	// plain exec.Command + HideWindow tended to leave mstsc hidden behind the
	// maximized framing uniTerm window ("double-click nothing happened").
	procShellExecute := windows.NewLazySystemDLL("shell32.dll").NewProc("ShellExecuteW")
	op, _ := windows.UTF16PtrFromString("open")
	file, _ := windows.UTF16PtrFromString(cmd)
	params, _ := windows.UTF16PtrFromString(path)
	ret, _, _ := procShellExecute.Call(
		0,
		uintptr(unsafe.Pointer(op)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(params)),
		0,
		1, // SW_SHOWNORMAL
	)
	if ret <= 32 {
		os.Remove(path)
		return fmt.Errorf("ShellExecute mstsc failed: %d", int(ret))
	}

	// mstsc reads the profile during startup; give it a beat, then clear the
	// temp file so the (weakly obfuscated) password does not linger on disk.
	go func() {
		time.Sleep(20 * time.Second)
		os.Remove(path)
	}()

	log.Writef("[RDP] launched mstsc profile %s for %s", path, config.Host)
	return nil
}