package apparmor

import (
	// #cgo LDFLAGS: -lapparmor
	// #include "./apparmor.h"
	"C"
)
import (
	"runtime"
	"syscall"
	"unsafe"
)

func ChangeHat(subprofile string, magicToken uint64) error {
	var ret uintptr
	if subprofile != "" {
		subProfileC := C.CString(subprofile)
		defer C.free(unsafe.Pointer(subProfileC))
		ret = uintptr(C.go_aa_change_hat(subProfileC, C.ulong(magicToken)))
	} else {
		ret = uintptr(C.go_aa_change_hat(nil, C.ulong(magicToken)))
	}

	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

func ExecuteInHat(subprofile string, fn func(), lockThread bool) error {
	if subprofile == "" {
		fn()
		return nil
	}
	if lockThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}
	token, err := GetMagicToken()
	if err != nil {
		return err
	}
	if err := ChangeHat(subprofile, token); err != nil {
		return err
	}
	fn()
	return ChangeHat("", token)
}

func ChangeProfile(subprofile string) error {
	var ret uintptr

	if subprofile != "" {
		subProfileC := C.CString(subprofile)
		defer C.free(unsafe.Pointer(subProfileC))
		ret = uintptr(C.go_aa_change_profile(subProfileC))
	} else {
		ret = uintptr(C.go_aa_change_profile(nil))
	}

	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}
