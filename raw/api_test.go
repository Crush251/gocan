package raw

import (
	"testing"
	"unsafe"
)

// TestConstants_PCANStatus 校验错误码常量值是否与 PCANBasic.h 完全一致。
//
// 若 PCAN 头文件升级导致值变化，这里会立刻失败 —— 这是有意为之，
// 因为高层 errors.go 的 Is/Has 都依赖这些位掩码值。
func TestConstants_PCANStatus(t *testing.T) {
	cases := []struct {
		name string
		got  TPCANStatus
		want uint32
	}{
		{"OK", PCAN_ERROR_OK, 0x00000},
		{"XMTFULL", PCAN_ERROR_XMTFULL, 0x00001},
		{"BUSLIGHT", PCAN_ERROR_BUSLIGHT, 0x00004},
		{"BUSHEAVY", PCAN_ERROR_BUSHEAVY, 0x00008},
		{"BUSOFF", PCAN_ERROR_BUSOFF, 0x00010},
		{"QRCVEMPTY", PCAN_ERROR_QRCVEMPTY, 0x00020},
		{"QOVERRUN", PCAN_ERROR_QOVERRUN, 0x00040},
		{"QXMTFULL", PCAN_ERROR_QXMTFULL, 0x00080},
		{"BUSPASSIVE", PCAN_ERROR_BUSPASSIVE, 0x40000},
		{"NODRIVER", PCAN_ERROR_NODRIVER, 0x00200},
		{"ILLPARAMTYPE", PCAN_ERROR_ILLPARAMTYPE, 0x04000},
		{"ILLPARAMVAL", PCAN_ERROR_ILLPARAMVAL, 0x08000},
		{"ILLOPERATION", PCAN_ERROR_ILLOPERATION, 0x80000},
		{"INITIALIZE", PCAN_ERROR_INITIALIZE, 0x4000000},
	}
	for _, c := range cases {
		if uint32(c.got) != c.want {
			t.Errorf("%s = 0x%X, want 0x%X", c.name, uint32(c.got), c.want)
		}
	}
}

func TestConstants_USBChannels(t *testing.T) {
	if PCAN_USBBUS1 != 0x51 {
		t.Errorf("PCAN_USBBUS1 = 0x%X, want 0x51", PCAN_USBBUS1)
	}
	if PCAN_USBBUS8 != 0x58 {
		t.Errorf("PCAN_USBBUS8 = 0x%X, want 0x58", PCAN_USBBUS8)
	}
	if PCAN_USBBUS9 != 0x509 {
		t.Errorf("PCAN_USBBUS9 = 0x%X, want 0x509", PCAN_USBBUS9)
	}
	if PCAN_USBBUS16 != 0x510 {
		t.Errorf("PCAN_USBBUS16 = 0x%X, want 0x510", PCAN_USBBUS16)
	}
}

func TestConstants_Baudrate(t *testing.T) {
	if PCAN_BAUD_1M != 0x0014 {
		t.Errorf("PCAN_BAUD_1M = 0x%X, want 0x0014", PCAN_BAUD_1M)
	}
	if PCAN_BAUD_500K != 0x001C {
		t.Errorf("PCAN_BAUD_500K = 0x%X, want 0x001C", PCAN_BAUD_500K)
	}
	if PCAN_BAUD_125K != 0x031C {
		t.Errorf("PCAN_BAUD_125K = 0x%X, want 0x031C", PCAN_BAUD_125K)
	}
}

// TPCANMsg 字段偏移必须严格与 C 结构匹配，否则 PCANBasic.dll 写入数据时会错位。
func TestTPCANMsg_FieldOffsets(t *testing.T) {
	var m TPCANMsg
	if got := unsafe.Offsetof(m.ID); got != 0 {
		t.Errorf("ID offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(m.MsgType); got != 4 {
		t.Errorf("MsgType offset = %d, want 4", got)
	}
	if got := unsafe.Offsetof(m.Len); got != 5 {
		t.Errorf("Len offset = %d, want 5", got)
	}
	if got := unsafe.Offsetof(m.Data); got != 6 {
		t.Errorf("Data offset = %d, want 6", got)
	}
}

func TestTPCANMsgFD_FieldOffsets(t *testing.T) {
	var m TPCANMsgFD
	if got := unsafe.Offsetof(m.ID); got != 0 {
		t.Errorf("ID offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(m.MsgType); got != 4 {
		t.Errorf("MsgType offset = %d, want 4", got)
	}
	if got := unsafe.Offsetof(m.DLC); got != 5 {
		t.Errorf("DLC offset = %d, want 5", got)
	}
	if got := unsafe.Offsetof(m.Data); got != 6 {
		t.Errorf("Data offset = %d, want 6", got)
	}
}

func TestTPCANTimestamp_FieldOffsets(t *testing.T) {
	var ts TPCANTimestamp
	if got := unsafe.Offsetof(ts.Millis); got != 0 {
		t.Errorf("Millis offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(ts.MillisOverflow); got != 4 {
		t.Errorf("MillisOverflow offset = %d, want 4", got)
	}
	if got := unsafe.Offsetof(ts.Micros); got != 6 {
		t.Errorf("Micros offset = %d, want 6", got)
	}
}

// 验证非 Windows 平台桩函数返回 PCAN_ERROR_ILLOPERATION。
// 在 Windows 上由于 DLL 可能不存在，会返回 PCAN_ERROR_NODRIVER；
// 两种情况都是预期"不可用"，我们只接受这两种状态码之一。
func TestStubReturns_OnNoHardware(t *testing.T) {
	status := Initialize(PCAN_USBBUS1, PCAN_BAUD_1M)
	if status != PCAN_ERROR_ILLOPERATION && status != PCAN_ERROR_NODRIVER {
		t.Errorf("Initialize without hardware/DLL returned 0x%X, want ILLOPERATION or NODRIVER",
			uint32(status))
	}
	_ = Uninitialize(PCAN_USBBUS1)
}

func TestEnsureLoaded_DoesNotPanic(t *testing.T) {
	// 在 Linux 上总返回 nil；Windows 上无 DLL 时返回 *fmt.wrapError。
	// 测试仅校验不 panic、可重入。
	_ = EnsureLoaded()
	_ = EnsureLoaded()
}
