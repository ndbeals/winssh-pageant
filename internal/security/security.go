package security

import (
	"golang.org/x/sys/windows"
)

func GetUserSID() (*windows.SID, error) {
	token := windows.GetCurrentProcessToken()
	user, err := token.GetTokenUser()
	if err != nil {
		return nil, err
	}

	return user.User.Sid, nil
}

func GetHandleSID(h windows.Handle) (*windows.SID, error) {
	securityDescriptor, err := windows.GetSecurityInfo(h, windows.SE_KERNEL_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return nil, err
	}

	sid, _, err := securityDescriptor.Owner()
	if err != nil {
		return nil, err
	}

	return sid, nil
}

func GetDefaultSID() (*windows.SID, error) {
	proc := windows.CurrentProcess()
	return GetHandleSID(proc)
}
