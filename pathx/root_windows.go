//go:build windows

package pathx

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type windowsRoot struct{ handle windows.Handle }

// Windows 使用 NtCreateFile 的 RootDirectory 约束每一段路径，OBJ_DONT_REPARSE
// 会使 junction、符号链接等 reparse point 在解析前被拒绝。
func openRoot(path string) (rootHandle, error) {
	if path == "" {
		return nil, ErrInvalidPath
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	handle, err := ntOpen(0, `\??\`+abs, windows.FILE_GENERIC_READ|windows.FILE_TRAVERSE, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE)
	if err != nil {
		return nil, secureWindowsError(err)
	}
	return &windowsRoot{handle: handle}, nil
}

func (r *windowsRoot) close() error { return windows.CloseHandle(r.handle) }

func (r *windowsRoot) openFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	parent, base, err := r.parent(name)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(parent)
	handle, err := ntOpen(parent, base, windowsAccess(flag), windowsDisposition(flag), 0)
	if err != nil {
		return nil, secureWindowsError(err)
	}
	return os.NewFile(uintptr(handle), base), nil
}

func (r *windowsRoot) mkdirAll(name string, perm fs.FileMode) error {
	parts, err := rootParts(name)
	if err != nil {
		return err
	}
	current, err := r.duplicate()
	if err != nil {
		return err
	}
	defer windows.CloseHandle(current)
	for _, part := range parts {
		next, err := ntOpen(current, part, windows.FILE_GENERIC_READ|windows.FILE_TRAVERSE, windows.FILE_OPEN_IF, windows.FILE_DIRECTORY_FILE)
		if err != nil {
			return secureWindowsError(err)
		}
		if err := windows.CloseHandle(current); err != nil {
			windows.CloseHandle(next)
			return err
		}
		current = next
	}
	return nil
}

func (r *windowsRoot) remove(name string) error {
	parent, base, err := r.parent(name)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(parent)
	handle, err := ntOpen(parent, base, windows.DELETE|windows.SYNCHRONIZE, windows.FILE_OPEN, 0)
	if err != nil {
		return secureWindowsError(err)
	}
	defer windows.CloseHandle(handle)
	var iosb windows.IO_STATUS_BLOCK
	deleteFile := byte(1)
	if err := windows.NtSetInformationFile(handle, &iosb, &deleteFile, 1, windows.FileDispositionInformation); err != nil {
		return secureWindowsError(err)
	}
	return nil
}

func (r *windowsRoot) rename(oldName, newName string) error {
	oldParent, oldBase, err := r.parent(oldName)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(oldParent)
	newParent, newBase, err := r.parent(newName)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(newParent)
	source, err := ntOpen(oldParent, oldBase, windows.DELETE|windows.SYNCHRONIZE, windows.FILE_OPEN, 0)
	if err != nil {
		return secureWindowsError(err)
	}
	defer windows.CloseHandle(source)
	return renameHandle(source, newParent, newBase)
}

func (r *windowsRoot) parent(name string) (windows.Handle, string, error) {
	parts, err := rootParts(name)
	if err != nil {
		return 0, "", err
	}
	current, err := r.duplicate()
	if err != nil {
		return 0, "", err
	}
	for _, part := range parts[:len(parts)-1] {
		next, err := ntOpen(current, part, windows.FILE_GENERIC_READ|windows.FILE_TRAVERSE, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE)
		if err != nil {
			windows.CloseHandle(current)
			return 0, "", secureWindowsError(err)
		}
		if err := windows.CloseHandle(current); err != nil {
			windows.CloseHandle(next)
			return 0, "", err
		}
		current = next
	}
	return current, parts[len(parts)-1], nil
}

func (r *windowsRoot) duplicate() (windows.Handle, error) {
	var duplicate windows.Handle
	process := windows.CurrentProcess()
	if err := windows.DuplicateHandle(process, r.handle, process, &duplicate, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return 0, err
	}
	return duplicate, nil
}

func ntOpen(parent windows.Handle, name string, access uint32, disposition uint32, options uint32) (windows.Handle, error) {
	objectName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return 0, err
	}
	oa := &windows.OBJECT_ATTRIBUTES{
		Length:        uint32(unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{})),
		RootDirectory: parent,
		ObjectName:    objectName,
		Attributes:    windows.OBJ_CASE_INSENSITIVE | windows.OBJ_DONT_REPARSE,
	}
	var handle windows.Handle
	var iosb windows.IO_STATUS_BLOCK
	var allocation int64
	err = windows.NtCreateFile(&handle, access, oa, &iosb, &allocation, 0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		disposition, options|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	if err != nil {
		return 0, err
	}
	return handle, nil
}

func windowsAccess(flag int) uint32 {
	if flag&os.O_RDWR != 0 {
		return windows.FILE_GENERIC_READ | windows.FILE_GENERIC_WRITE
	}
	if flag&os.O_WRONLY != 0 {
		return windows.FILE_GENERIC_WRITE
	}
	return windows.FILE_GENERIC_READ
}

func windowsDisposition(flag int) uint32 {
	switch {
	case flag&os.O_CREATE != 0 && flag&os.O_EXCL != 0:
		return windows.FILE_CREATE
	case flag&os.O_CREATE != 0 && flag&os.O_TRUNC != 0:
		return windows.FILE_OVERWRITE_IF
	case flag&os.O_CREATE != 0:
		return windows.FILE_OPEN_IF
	case flag&os.O_TRUNC != 0:
		return windows.FILE_OVERWRITE
	default:
		return windows.FILE_OPEN
	}
}

type fileRenameInformation struct {
	ReplaceIfExists uint32
	RootDirectory   windows.Handle
	FileNameLength  uint32
	FileName        [1]uint16
}

func renameHandle(source, destinationRoot windows.Handle, destination string) error {
	encoded, err := windows.UTF16FromString(destination)
	if err != nil {
		return err
	}
	nameBytes := (len(encoded) - 1) * 2
	var header fileRenameInformation
	size := int(unsafe.Offsetof(header.FileName)) + nameBytes
	buffer := make([]byte, size)
	info := (*fileRenameInformation)(unsafe.Pointer(&buffer[0]))
	info.ReplaceIfExists = windows.FILE_RENAME_REPLACE_IF_EXISTS
	info.RootDirectory = destinationRoot
	info.FileNameLength = uint32(nameBytes)
	// 缓冲区按目标名称实际长度分配，不能使用 MAX_LONG_PATH 固定数组，
	// 否则启用长路径的 Windows 环境会在重命名时发生越界访问。
	fileName := unsafe.Slice((*uint16)(unsafe.Pointer(&info.FileName[0])), len(encoded)-1)
	copy(fileName, encoded[:len(encoded)-1])
	var iosb windows.IO_STATUS_BLOCK
	if err := windows.NtSetInformationFile(source, &iosb, &buffer[0], uint32(len(buffer)), windows.FileRenameInformation); err != nil {
		return secureWindowsError(err)
	}
	return nil
}

func rootParts(name string) ([]string, error) {
	clean, err := SanitizeRelativePath(name)
	if err != nil {
		return nil, err
	}
	return strings.Split(clean, "/"), nil
}

func secureWindowsError(err error) error {
	if err == nil {
		return nil
	}
	var status windows.NTStatus
	if errors.As(err, &status) {
		switch status {
		case windows.STATUS_REPARSE_POINT_ENCOUNTERED, windows.STATUS_STOPPED_ON_SYMLINK:
			return fmt.Errorf("%w: reparse point is not allowed", ErrPathEscape)
		}
	}
	return err
}
