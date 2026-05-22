package raw

// 与 PCANBasic.h 对应的基础类型。
type (
	TPCANHandle      uint16
	TPCANStatus      uint32
	TPCANBaudrate    uint16
	TPCANType        uint8
	TPCANMessageType uint8
	TPCANParameter   uint8
)

// TPCANTimestampFD 是 CAN FD 用的微秒时间戳类型（直接是 uint64，无需包装结构）。
type TPCANTimestampFD = uint64

// TPCANMsg 字段顺序与 C 结构 TPCANMsg 一致。
//
// 注意：因 Go 自然对齐规则，unsafe.Sizeof(TPCANMsg{}) 可能为 16 而非 C 的 14。
// 这不影响 PCANBasic 调用 —— DLL 仅访问前 14 字节，尾部 padding 是 Go 自有内存，
// 对端不感知。
type TPCANMsg struct {
	ID      uint32           // CAN 标识符（11 位或 29 位）
	MsgType TPCANMessageType // 帧类型位组合（STANDARD/EXTENDED/RTR/...）
	Len     uint8            // 数据长度，0..8
	Data    [8]byte          // 数据
}

// TPCANMsgFD 与 C 结构 TPCANMsgFD 字段对应。
//
// 注意：DLC 字段是 CAN FD 协议的"长度码"，离散值 0..15。
// 0..8 直接表示字节数；9..15 分别表示 12/16/20/24/32/48/64 字节。
type TPCANMsgFD struct {
	ID      uint32           // CAN 标识符
	MsgType TPCANMessageType // FD/BRS/ESI/EXTENDED/... 位组合
	DLC     uint8            // 长度码 0..15
	Data    [64]byte         // 数据，最长 64 字节
}

// TPCANTimestamp 与 C 结构 TPCANTimestamp 字段对应（Classical CAN 用）。
type TPCANTimestamp struct {
	Millis         uint32 // 毫秒部分
	MillisOverflow uint16 // millis 溢出次数
	Micros         uint16 // 微秒部分（0..999）
}
