//go:build windows

package session

import (
	"strings"
	"syscall"
	"unsafe"

	"github.com/ys-ll/uniterm/backend/log"
	"golang.org/x/sys/windows"
)

// RDP auto-logon under NLA/CredSSP.
//
// The RDP ActiveX control ignores IMsTscNonScriptable.ClearTextPassword when
// Network Level Authentication (CredSSP) is negotiated, so the control still
// shows the Windows credential prompt even though the password is known. This
// is a well-documented limitation of the control.
//
// The reliable way a saved RDP password auto-logs on under NLA — the same
// mechanism mstsc's "Remember credentials" and other RDP clients use — is to
// store the credential in the Windows Credential Manager under the target
// "TERMSRV/<host>". CredSSP then retrieves it automatically at logon time.
// golang.org/x/sys does not export the Credential Manager API, so we call
// advapi32 directly.

// windowsCredential mirrors Win32 CREDENTIALW. Field order/types match C so
// Go's natural ABI alignment reproduces the same 64-bit layout.
type windowsCredential struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        syscall.Filetime
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

const (
	credTypeGeneric         = 1 // CRED_TYPE_GENERIC
	credPersistLocalMachine = 3 // CRED_PERSIST_LOCAL_MACHINE
)

var (
	advapi32       = windows.NewLazySystemDLL("advapi32.dll")
	procCredWrite  = advapi32.NewProc("CredWriteW")
	procCredDelete = advapi32.NewProc("CredDeleteW")
)

// rdpCredentialTarget builds the Credential Manager target for an RDP host:
// the same "TERMSRV/<host>" namespace mstsc uses for saved RDP credentials.
func rdpCredentialTarget(host string) *uint16 {
	p, _ := windows.UTF16PtrFromString("TERMSRV/" + host)
	return p
}

// rdpCredentialUser returns the logon user shown to CredSSP, honoring an
// explicit RdpDomain and the "DOMAIN\user" form of the username.
func rdpCredentialUser(domain, user string) string {
	if domain != "" {
		return domain + `\` + user
	}
	if strings.Contains(user, `\`) || strings.Contains(user, "@") {
		return user
	}
	return user
}

// rdpSaveCredential writes a generic "TERMSRV/<host>" credential so CredSSP
// auto-logs on under NLA. Returns the OS error on failure so the caller can
// log it (best-effort — the connect path still proceeds).
func rdpSaveCredential(host, domain, user, password string) error {
	blob := syscall.StringToUTF16(password)
	var blobPtr *byte
	if len(blob) > 0 {
		blobPtr = (*byte)(unsafe.Pointer(&blob[0]))
	}

	userName, _ := windows.UTF16PtrFromString(rdpCredentialUser(domain, user))

	c := &windowsCredential{
		Type:               credTypeGeneric,
		TargetName:         rdpCredentialTarget(host),
		Comment:            windows.StringToUTF16Ptr("uniTerm RDP credential"),
		CredentialBlobSize: uint32(len(blob) * 2),
		CredentialBlob:     blobPtr,
		Persist:            credPersistLocalMachine,
		UserName:           userName,
	}

	ret, _, _ := procCredWrite.Call(uintptr(unsafe.Pointer(c)), 0)
	if ret == 0 {
		return windows.GetLastError()
	}
	return nil
}

// rdpDeleteCredential removes a previously saved "TERMSRV/<host>" credential
// (used when the stored password is cleared).
func rdpDeleteCredential(host string) {
	procCredDelete.Call(
		uintptr(unsafe.Pointer(rdpCredentialTarget(host))),
		uintptr(credTypeGeneric),
		0,
	)
}

// rdpSyncCredential writes or clears the Credential Manager entry for host to
// match the current stored password, so NLA auto-logon always uses the latest
// value. Failures are logged but never block the connect.
func rdpSyncCredential(host, domain, user, password string) {
	if password == "" {
		rdpDeleteCredential(host)
		return
	}
	if err := rdpSaveCredential(host, domain, user, password); err != nil {
		log.Writef("[RDP] save credential for TERMSRV/%s failed: %v", host, err)
	}
}