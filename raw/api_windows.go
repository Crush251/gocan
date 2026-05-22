//go:build windows

package raw

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	loadOnce sync.Once
	loadErr  error

	dll *windows.LazyDLL

	procInitialize     *windows.LazyProc
	procInitializeFD   *windows.LazyProc
	procUninitialize   *windows.LazyProc
	procReset          *windows.LazyProc
	procGetStatus      *windows.LazyProc
	procRead           *windows.LazyProc
	procReadFD         *windows.LazyProc
	procWrite          *windows.LazyProc
	procWriteFD        *windows.LazyProc
	procFilterMessages *windows.LazyProc
	procGetValue       *windows.LazyProc
	procSetValue       *windows.LazyProc
	procGetErrorText   *windows.LazyProc
)

// EnsureLoaded 显式触发 PCANBasic.dll 的加载。
//
// 默认路径 "PCANBasic.dll"（按 Windows 标准 DLL 搜索路径解析）；
// 可通过环境变量 PCANBASIC_DLL_PATH 覆盖为绝对/相对路径。
//
// 加载失败时所有后续 API 调用返回 PCAN_ERROR_NODRIVER。
// 同一进程内 sync.Once 只尝试一次。
func EnsureLoaded() error {
	loadOnce.Do(func() {
		path := os.Getenv("PCANBASIC_DLL_PATH")
		if path == "" {
			path = "PCANBasic.dll"
		}
		dll = windows.NewLazyDLL(path)
		if err := dll.Load(); err != nil {
			loadErr = fmt.Errorf("load PCANBasic dll %q: %w", path, err)
			return
		}
		procInitialize = dll.NewProc("CAN_Initialize")
		procInitializeFD = dll.NewProc("CAN_InitializeFD")
		procUninitialize = dll.NewProc("CAN_Uninitialize")
		procReset = dll.NewProc("CAN_Reset")
		procGetStatus = dll.NewProc("CAN_GetStatus")
		procRead = dll.NewProc("CAN_Read")
		procReadFD = dll.NewProc("CAN_ReadFD")
		procWrite = dll.NewProc("CAN_Write")
		procWriteFD = dll.NewProc("CAN_WriteFD")
		procFilterMessages = dll.NewProc("CAN_FilterMessages")
		procGetValue = dll.NewProc("CAN_GetValue")
		procSetValue = dll.NewProc("CAN_SetValue")
		procGetErrorText = dll.NewProc("CAN_GetErrorText")
	})
	return loadErr
}

func ensure() bool {
	return EnsureLoaded() == nil
}

// Initialize 是 CAN_Initialize 的 Go 绑定（Classical CAN）。
func Initialize(ch TPCANHandle, br TPCANBaudrate) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procInitialize.Call(uintptr(ch), uintptr(br), 0, 0, 0)
	return TPCANStatus(r)
}

// InitializeFD 是 CAN_InitializeFD 的 Go 绑定（CAN FD）。
// bitrateFD 是 PCAN 官方的位率字符串，例如：
// "f_clock=80000000, nom_brp=2, nom_tseg1=63, nom_tseg2=16, nom_sjw=16, data_brp=2, data_tseg1=15, data_tseg2=4, data_sjw=4"
func InitializeFD(ch TPCANHandle, bitrateFD string) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	b, err := windows.BytePtrFromString(bitrateFD)
	if err != nil {
		return PCAN_ERROR_ILLPARAMVAL
	}
	r, _, _ := procInitializeFD.Call(uintptr(ch), uintptr(unsafe.Pointer(b)))
	return TPCANStatus(r)
}

// Uninitialize 是 CAN_Uninitialize 的 Go 绑定。
func Uninitialize(ch TPCANHandle) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procUninitialize.Call(uintptr(ch))
	return TPCANStatus(r)
}

// Reset 是 CAN_Reset 的 Go 绑定。会清空 PCAN 内部的收发队列。
func Reset(ch TPCANHandle) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procReset.Call(uintptr(ch))
	return TPCANStatus(r)
}

// GetStatus 是 CAN_GetStatus 的 Go 绑定。返回值就是当前通道状态。
func GetStatus(ch TPCANHandle) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procGetStatus.Call(uintptr(ch))
	return TPCANStatus(r)
}

// Read 是 CAN_Read 的 Go 绑定（Classical CAN）。
// 队列空时返回 PCAN_ERROR_QRCVEMPTY，调用方应将其视为"没有新数据"而非真正错误。
func Read(ch TPCANHandle, m *TPCANMsg, t *TPCANTimestamp) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procRead.Call(uintptr(ch),
		uintptr(unsafe.Pointer(m)),
		uintptr(unsafe.Pointer(t)))
	return TPCANStatus(r)
}

// ReadFD 是 CAN_ReadFD 的 Go 绑定（CAN FD）。
func ReadFD(ch TPCANHandle, m *TPCANMsgFD, t *TPCANTimestampFD) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procReadFD.Call(uintptr(ch),
		uintptr(unsafe.Pointer(m)),
		uintptr(unsafe.Pointer(t)))
	return TPCANStatus(r)
}

// Write 是 CAN_Write 的 Go 绑定（Classical CAN）。
func Write(ch TPCANHandle, m *TPCANMsg) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procWrite.Call(uintptr(ch), uintptr(unsafe.Pointer(m)))
	return TPCANStatus(r)
}

// WriteFD 是 CAN_WriteFD 的 Go 绑定（CAN FD）。
func WriteFD(ch TPCANHandle, m *TPCANMsgFD) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procWriteFD.Call(uintptr(ch), uintptr(unsafe.Pointer(m)))
	return TPCANStatus(r)
}

// FilterMessages 是 CAN_FilterMessages 的 Go 绑定。
// 设置接收 ID 范围 [fromID, toID]，mode 区分 11/29 位 ID。
func FilterMessages(ch TPCANHandle, fromID, toID uint32, mode TPCANMessageType) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procFilterMessages.Call(
		uintptr(ch), uintptr(fromID), uintptr(toID), uintptr(mode))
	return TPCANStatus(r)
}

// GetValue 是 CAN_GetValue 的 Go 绑定。
// buf 指向接收缓冲（由调用方确保大小为 n）。
func GetValue(ch TPCANHandle, p TPCANParameter, buf unsafe.Pointer, n uint32) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procGetValue.Call(uintptr(ch), uintptr(p), uintptr(buf), uintptr(n))
	return TPCANStatus(r)
}

// SetValue 是 CAN_SetValue 的 Go 绑定。
func SetValue(ch TPCANHandle, p TPCANParameter, buf unsafe.Pointer, n uint32) TPCANStatus {
	if !ensure() {
		return PCAN_ERROR_NODRIVER
	}
	r, _, _ := procSetValue.Call(uintptr(ch), uintptr(p), uintptr(buf), uintptr(n))
	return TPCANStatus(r)
}

// GetErrorText 是 CAN_GetErrorText 的 Go 绑定。
// 返回 (描述文本, 调用状态)：仅当状态为 PCAN_ERROR_OK 时文本有效。
func GetErrorText(code TPCANStatus, lang uint16) (string, TPCANStatus) {
	if !ensure() {
		return "", PCAN_ERROR_NODRIVER
	}
	var buf [256]byte
	r, _, _ := procGetErrorText.Call(uintptr(code), uintptr(lang),
		uintptr(unsafe.Pointer(&buf[0])))
	status := TPCANStatus(r)
	if status != PCAN_ERROR_OK {
		return "", status
	}
	// C 字符串以 NUL 结尾，按 NUL 截断。
	n := 0
	for n < len(buf) && buf[n] != 0 {
		n++
	}
	return string(buf[:n]), PCAN_ERROR_OK
}
