//go:build windows

package antigravity

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const credTypeGeneric = 1

type windowsCredential struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        windows.Filetime
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

var (
	advapi32     = windows.NewLazySystemDLL("advapi32.dll")
	procCredRead = advapi32.NewProc("CredReadW")
	procCredFree = advapi32.NewProc("CredFree")
)

func readCLILoginPayload() ([]byte, error) {
	target, errTarget := windows.UTF16PtrFromString("gemini:antigravity")
	if errTarget != nil {
		return nil, fmt.Errorf("encode credential target: %w", errTarget)
	}
	var credential *windowsCredential
	ok, _, errCall := procCredRead.Call(uintptr(unsafe.Pointer(target)), credTypeGeneric, 0, uintptr(unsafe.Pointer(&credential)))
	if ok == 0 {
		if errors.Is(errCall, windows.ERROR_NOT_FOUND) {
			return nil, ErrCLILoginNotFound
		}
		return nil, fmt.Errorf("read antigravity cli login: %w", errCall)
	}
	defer func() { _, _, _ = procCredFree.Call(uintptr(unsafe.Pointer(credential))) }()
	if credential.CredentialBlobSize == 0 || credential.CredentialBlob == nil {
		return nil, ErrCLILoginNotFound
	}
	blob := unsafe.Slice(credential.CredentialBlob, credential.CredentialBlobSize)
	return append([]byte(nil), blob...), nil
}
