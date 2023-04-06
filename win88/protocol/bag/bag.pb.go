// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.19.4
// source: bag.proto

package bag

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

//操作结果
type OpResultCode int32

const (
	OpResultCode_OPRC_Sucess    OpResultCode = 0 //成功
	OpResultCode_OPRC_Error     OpResultCode = 1 //未知错误
	OpResultCode_OPRC_UseUp     OpResultCode = 2 //道具不足
	OpResultCode_OPRC_IdErr     OpResultCode = 3 //物品编号不存在
	OpResultCode_OPRC_DbErr     OpResultCode = 4 //存储出错
	OpResultCode_OPRC_BagFull   OpResultCode = 5 //背包已满
	OpResultCode_OPRC_NotPlayer OpResultCode = 6 //找不到玩家
)

// Enum value maps for OpResultCode.
var (
	OpResultCode_name = map[int32]string{
		0: "OPRC_Sucess",
		1: "OPRC_Error",
		2: "OPRC_UseUp",
		3: "OPRC_IdErr",
		4: "OPRC_DbErr",
		5: "OPRC_BagFull",
		6: "OPRC_NotPlayer",
	}
	OpResultCode_value = map[string]int32{
		"OPRC_Sucess":    0,
		"OPRC_Error":     1,
		"OPRC_UseUp":     2,
		"OPRC_IdErr":     3,
		"OPRC_DbErr":     4,
		"OPRC_BagFull":   5,
		"OPRC_NotPlayer": 6,
	}
)

func (x OpResultCode) Enum() *OpResultCode {
	p := new(OpResultCode)
	*p = x
	return p
}

func (x OpResultCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (OpResultCode) Descriptor() protoreflect.EnumDescriptor {
	return file_bag_proto_enumTypes[0].Descriptor()
}

func (OpResultCode) Type() protoreflect.EnumType {
	return &file_bag_proto_enumTypes[0]
}

func (x OpResultCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use OpResultCode.Descriptor instead.
func (OpResultCode) EnumDescriptor() ([]byte, []int) {
	return file_bag_proto_rawDescGZIP(), []int{0}
}

// 商城
type SPacketID int32

const (
	SPacketID_PACKET_BAG_ZERO       SPacketID = 0    // 弃用消息号
	SPacketID_PACKET_ALL_BAG_INFO   SPacketID = 2530 //请求背包数据
	SPacketID_PACKET_ALL_BAG_USE    SPacketID = 2531 //使用背包道具
	SPacketID_PACKET_SC_SYNCBAGDATA SPacketID = 2532 //背包数据更新
	SPacketID_PACKET_ALL_BAG_END    SPacketID = 2549 //最大消息号
)

// Enum value maps for SPacketID.
var (
	SPacketID_name = map[int32]string{
		0:    "PACKET_BAG_ZERO",
		2530: "PACKET_ALL_BAG_INFO",
		2531: "PACKET_ALL_BAG_USE",
		2532: "PACKET_SC_SYNCBAGDATA",
		2549: "PACKET_ALL_BAG_END",
	}
	SPacketID_value = map[string]int32{
		"PACKET_BAG_ZERO":       0,
		"PACKET_ALL_BAG_INFO":   2530,
		"PACKET_ALL_BAG_USE":    2531,
		"PACKET_SC_SYNCBAGDATA": 2532,
		"PACKET_ALL_BAG_END":    2549,
	}
)

func (x SPacketID) Enum() *SPacketID {
	p := new(SPacketID)
	*p = x
	return p
}

func (x SPacketID) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SPacketID) Descriptor() protoreflect.EnumDescriptor {
	return file_bag_proto_enumTypes[1].Descriptor()
}

func (SPacketID) Type() protoreflect.EnumType {
	return &file_bag_proto_enumTypes[1]
}

func (x SPacketID) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SPacketID.Descriptor instead.
func (SPacketID) EnumDescriptor() ([]byte, []int) {
	return file_bag_proto_rawDescGZIP(), []int{1}
}

//物品信息 后续精简
type ItemInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//数据表数据
	ItemId  int32 `protobuf:"varint,1,opt,name=ItemId,proto3" json:"ItemId,omitempty"`   // 物品ID
	ItemNum int32 `protobuf:"varint,2,opt,name=ItemNum,proto3" json:"ItemNum,omitempty"` // 物品数量
	//  string Name = 3; // 名称
	//  repeated int32 ShowLocation = 4; // 分页类型 1，道具类 	2，资源类	3，兑换类
	//  repeated int32 Classify = 5; // 分页类型 1，道具类 	2，资源类	3，兑换类
	//  int32 Type = 6; // 道具种类 1，宠物碎片 2，角色碎片
	//  repeated int32 Effect0 = 7; // 竖版道具功能 1，使用 2，赠送 3，出售
	//  repeated int32 Effect = 8; // 横版道具功能 1，使用 2，赠送 3，出售
	//  int32 SaleType = 9; // 出售类型
	//  int32 SaleGold = 10; // 出售金额
	//  int32 Composition = 11;  // 能否叠加 1，能 2，不能
	//  int32 CompositionMax = 12; // 叠加上限
	//  int32 Time = 13; // 道具时效 0为永久
	//  string Location = 14; // 跳转页面
	//  string Describe = 15;  // 道具描述
	ObtainTime int64 `protobuf:"varint,3,opt,name=ObtainTime,proto3" json:"ObtainTime,omitempty"` //获取的时间
}

func (x *ItemInfo) Reset() {
	*x = ItemInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bag_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ItemInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ItemInfo) ProtoMessage() {}

func (x *ItemInfo) ProtoReflect() protoreflect.Message {
	mi := &file_bag_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ItemInfo.ProtoReflect.Descriptor instead.
func (*ItemInfo) Descriptor() ([]byte, []int) {
	return file_bag_proto_rawDescGZIP(), []int{0}
}

func (x *ItemInfo) GetItemId() int32 {
	if x != nil {
		return x.ItemId
	}
	return 0
}

func (x *ItemInfo) GetItemNum() int32 {
	if x != nil {
		return x.ItemNum
	}
	return 0
}

func (x *ItemInfo) GetObtainTime() int64 {
	if x != nil {
		return x.ObtainTime
	}
	return 0
}

//PACKET_ALL_BAG_INFO
type CSBagInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NowLocation int32 `protobuf:"varint,1,opt,name=NowLocation,proto3" json:"NowLocation,omitempty"` //0.通用 1.大厅 2.Tienlen 3.捕鱼
}

func (x *CSBagInfo) Reset() {
	*x = CSBagInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bag_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CSBagInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CSBagInfo) ProtoMessage() {}

func (x *CSBagInfo) ProtoReflect() protoreflect.Message {
	mi := &file_bag_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CSBagInfo.ProtoReflect.Descriptor instead.
func (*CSBagInfo) Descriptor() ([]byte, []int) {
	return file_bag_proto_rawDescGZIP(), []int{1}
}

func (x *CSBagInfo) GetNowLocation() int32 {
	if x != nil {
		return x.NowLocation
	}
	return 0
}

//PACKET_ALL_BAG_INFO
type SCBagInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RetCode   OpResultCode `protobuf:"varint,1,opt,name=RetCode,proto3,enum=bag.OpResultCode" json:"RetCode,omitempty"`
	Infos     []*ItemInfo  `protobuf:"bytes,2,rep,name=Infos,proto3" json:"Infos,omitempty"`          // 商品信息
	BagNumMax int32        `protobuf:"varint,3,opt,name=BagNumMax,proto3" json:"BagNumMax,omitempty"` //最大格子
}

func (x *SCBagInfo) Reset() {
	*x = SCBagInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bag_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SCBagInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SCBagInfo) ProtoMessage() {}

func (x *SCBagInfo) ProtoReflect() protoreflect.Message {
	mi := &file_bag_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SCBagInfo.ProtoReflect.Descriptor instead.
func (*SCBagInfo) Descriptor() ([]byte, []int) {
	return file_bag_proto_rawDescGZIP(), []int{2}
}

func (x *SCBagInfo) GetRetCode() OpResultCode {
	if x != nil {
		return x.RetCode
	}
	return OpResultCode_OPRC_Sucess
}

func (x *SCBagInfo) GetInfos() []*ItemInfo {
	if x != nil {
		return x.Infos
	}
	return nil
}

func (x *SCBagInfo) GetBagNumMax() int32 {
	if x != nil {
		return x.BagNumMax
	}
	return 0
}

//PACKET_ALL_BAG_USE
type CSUpBagInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ItemId     int32 `protobuf:"varint,1,opt,name=ItemId,proto3" json:"ItemId,omitempty"`         //物品ID
	ItemNum    int32 `protobuf:"varint,2,opt,name=ItemNum,proto3" json:"ItemNum,omitempty"`       //物品数量
	Opt        int32 `protobuf:"varint,3,opt,name=Opt,proto3" json:"Opt,omitempty"`               //操作 0.使用 1.赠送 2.出售
	AcceptSnId int32 `protobuf:"varint,4,opt,name=AcceptSnId,proto3" json:"AcceptSnId,omitempty"` //被赠送玩家id
	NowEffect  int32 `protobuf:"varint,5,opt,name=NowEffect,proto3" json:"NowEffect,omitempty"`   //0.竖版 1.横版
	ShowId     int64 `protobuf:"varint,6,opt,name=ShowId,proto3" json:"ShowId,omitempty"`         // 邮件显示位置  0 所有大厅都显示 1 主大厅显示 2 len大厅显示 4 fish大厅显示
}

func (x *CSUpBagInfo) Reset() {
	*x = CSUpBagInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bag_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CSUpBagInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CSUpBagInfo) ProtoMessage() {}

func (x *CSUpBagInfo) ProtoReflect() protoreflect.Message {
	mi := &file_bag_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CSUpBagInfo.ProtoReflect.Descriptor instead.
func (*CSUpBagInfo) Descriptor() ([]byte, []int) {
	return file_bag_proto_rawDescGZIP(), []int{3}
}

func (x *CSUpBagInfo) GetItemId() int32 {
	if x != nil {
		return x.ItemId
	}
	return 0
}

func (x *CSUpBagInfo) GetItemNum() int32 {
	if x != nil {
		return x.ItemNum
	}
	return 0
}

func (x *CSUpBagInfo) GetOpt() int32 {
	if x != nil {
		return x.Opt
	}
	return 0
}

func (x *CSUpBagInfo) GetAcceptSnId() int32 {
	if x != nil {
		return x.AcceptSnId
	}
	return 0
}

func (x *CSUpBagInfo) GetNowEffect() int32 {
	if x != nil {
		return x.NowEffect
	}
	return 0
}

func (x *CSUpBagInfo) GetShowId() int64 {
	if x != nil {
		return x.ShowId
	}
	return 0
}

//PACKET_ALL_BAG_USE
type SCUpBagInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RetCode    OpResultCode `protobuf:"varint,1,opt,name=RetCode,proto3,enum=bag.OpResultCode" json:"RetCode,omitempty"`
	NowItemId  int32        `protobuf:"varint,2,opt,name=NowItemId,proto3" json:"NowItemId,omitempty"`   //当前物品物品ID
	NowItemNum int32        `protobuf:"varint,3,opt,name=NowItemNum,proto3" json:"NowItemNum,omitempty"` //当前物品剩余数量
	//使用道具获得的
	Coin    int64 `protobuf:"varint,4,opt,name=Coin,proto3" json:"Coin,omitempty"`       //金币
	Diamond int64 `protobuf:"varint,5,opt,name=Diamond,proto3" json:"Diamond,omitempty"` // 钻石
	//使用道具获得的
	Infos []*ItemInfo `protobuf:"bytes,6,rep,name=Infos,proto3" json:"Infos,omitempty"` // 物品信息
}

func (x *SCUpBagInfo) Reset() {
	*x = SCUpBagInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bag_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SCUpBagInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SCUpBagInfo) ProtoMessage() {}

func (x *SCUpBagInfo) ProtoReflect() protoreflect.Message {
	mi := &file_bag_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SCUpBagInfo.ProtoReflect.Descriptor instead.
func (*SCUpBagInfo) Descriptor() ([]byte, []int) {
	return file_bag_proto_rawDescGZIP(), []int{4}
}

func (x *SCUpBagInfo) GetRetCode() OpResultCode {
	if x != nil {
		return x.RetCode
	}
	return OpResultCode_OPRC_Sucess
}

func (x *SCUpBagInfo) GetNowItemId() int32 {
	if x != nil {
		return x.NowItemId
	}
	return 0
}

func (x *SCUpBagInfo) GetNowItemNum() int32 {
	if x != nil {
		return x.NowItemNum
	}
	return 0
}

func (x *SCUpBagInfo) GetCoin() int64 {
	if x != nil {
		return x.Coin
	}
	return 0
}

func (x *SCUpBagInfo) GetDiamond() int64 {
	if x != nil {
		return x.Diamond
	}
	return 0
}

func (x *SCUpBagInfo) GetInfos() []*ItemInfo {
	if x != nil {
		return x.Infos
	}
	return nil
}

//PACKET_SC_SYNCBAGDATA
type SCSyncBagData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Infos []*ItemInfo `protobuf:"bytes,1,rep,name=Infos,proto3" json:"Infos,omitempty"` // 物品信息
}

func (x *SCSyncBagData) Reset() {
	*x = SCSyncBagData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bag_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SCSyncBagData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SCSyncBagData) ProtoMessage() {}

func (x *SCSyncBagData) ProtoReflect() protoreflect.Message {
	mi := &file_bag_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SCSyncBagData.ProtoReflect.Descriptor instead.
func (*SCSyncBagData) Descriptor() ([]byte, []int) {
	return file_bag_proto_rawDescGZIP(), []int{5}
}

func (x *SCSyncBagData) GetInfos() []*ItemInfo {
	if x != nil {
		return x.Infos
	}
	return nil
}

var File_bag_proto protoreflect.FileDescriptor

var file_bag_proto_rawDesc = []byte{
	0x0a, 0x09, 0x62, 0x61, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x62, 0x61, 0x67,
	0x22, 0x5c, 0x0a, 0x08, 0x49, 0x74, 0x65, 0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x16, 0x0a, 0x06,
	0x49, 0x74, 0x65, 0x6d, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x49, 0x74,
	0x65, 0x6d, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x49, 0x74, 0x65, 0x6d, 0x4e, 0x75, 0x6d, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x49, 0x74, 0x65, 0x6d, 0x4e, 0x75, 0x6d, 0x12, 0x1e,
	0x0a, 0x0a, 0x4f, 0x62, 0x74, 0x61, 0x69, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x0a, 0x4f, 0x62, 0x74, 0x61, 0x69, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x22, 0x2d,
	0x0a, 0x09, 0x43, 0x53, 0x42, 0x61, 0x67, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x20, 0x0a, 0x0b, 0x4e,
	0x6f, 0x77, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0b, 0x4e, 0x6f, 0x77, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x7b, 0x0a,
	0x09, 0x53, 0x43, 0x42, 0x61, 0x67, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x2b, 0x0a, 0x07, 0x52, 0x65,
	0x74, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x11, 0x2e, 0x62, 0x61,
	0x67, 0x2e, 0x4f, 0x70, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x07,
	0x52, 0x65, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x23, 0x0a, 0x05, 0x49, 0x6e, 0x66, 0x6f, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x62, 0x61, 0x67, 0x2e, 0x49, 0x74, 0x65,
	0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x05, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x12, 0x1c, 0x0a, 0x09,
	0x42, 0x61, 0x67, 0x4e, 0x75, 0x6d, 0x4d, 0x61, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x09, 0x42, 0x61, 0x67, 0x4e, 0x75, 0x6d, 0x4d, 0x61, 0x78, 0x22, 0xa7, 0x01, 0x0a, 0x0b, 0x43,
	0x53, 0x55, 0x70, 0x42, 0x61, 0x67, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x49, 0x74,
	0x65, 0x6d, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x49, 0x74, 0x65, 0x6d,
	0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x49, 0x74, 0x65, 0x6d, 0x4e, 0x75, 0x6d, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x07, 0x49, 0x74, 0x65, 0x6d, 0x4e, 0x75, 0x6d, 0x12, 0x10, 0x0a, 0x03,
	0x4f, 0x70, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x4f, 0x70, 0x74, 0x12, 0x1e,
	0x0a, 0x0a, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x53, 0x6e, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0a, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x53, 0x6e, 0x49, 0x64, 0x12, 0x1c,
	0x0a, 0x09, 0x4e, 0x6f, 0x77, 0x45, 0x66, 0x66, 0x65, 0x63, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x09, 0x4e, 0x6f, 0x77, 0x45, 0x66, 0x66, 0x65, 0x63, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x53, 0x68, 0x6f, 0x77, 0x49, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x53, 0x68,
	0x6f, 0x77, 0x49, 0x64, 0x22, 0xcb, 0x01, 0x0a, 0x0b, 0x53, 0x43, 0x55, 0x70, 0x42, 0x61, 0x67,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x2b, 0x0a, 0x07, 0x52, 0x65, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x11, 0x2e, 0x62, 0x61, 0x67, 0x2e, 0x4f, 0x70, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x07, 0x52, 0x65, 0x74, 0x43, 0x6f, 0x64,
	0x65, 0x12, 0x1c, 0x0a, 0x09, 0x4e, 0x6f, 0x77, 0x49, 0x74, 0x65, 0x6d, 0x49, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x4e, 0x6f, 0x77, 0x49, 0x74, 0x65, 0x6d, 0x49, 0x64, 0x12,
	0x1e, 0x0a, 0x0a, 0x4e, 0x6f, 0x77, 0x49, 0x74, 0x65, 0x6d, 0x4e, 0x75, 0x6d, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0a, 0x4e, 0x6f, 0x77, 0x49, 0x74, 0x65, 0x6d, 0x4e, 0x75, 0x6d, 0x12,
	0x12, 0x0a, 0x04, 0x43, 0x6f, 0x69, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x43,
	0x6f, 0x69, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x44, 0x69, 0x61, 0x6d, 0x6f, 0x6e, 0x64, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x44, 0x69, 0x61, 0x6d, 0x6f, 0x6e, 0x64, 0x12, 0x23, 0x0a,
	0x05, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x62,
	0x61, 0x67, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x05, 0x49, 0x6e, 0x66,
	0x6f, 0x73, 0x22, 0x34, 0x0a, 0x0d, 0x53, 0x43, 0x53, 0x79, 0x6e, 0x63, 0x42, 0x61, 0x67, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x23, 0x0a, 0x05, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x62, 0x61, 0x67, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x05, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x2a, 0x85, 0x01, 0x0a, 0x0c, 0x4f, 0x70, 0x52,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x0f, 0x0a, 0x0b, 0x4f, 0x50, 0x52,
	0x43, 0x5f, 0x53, 0x75, 0x63, 0x65, 0x73, 0x73, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x4f, 0x50,
	0x52, 0x43, 0x5f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x4f, 0x50,
	0x52, 0x43, 0x5f, 0x55, 0x73, 0x65, 0x55, 0x70, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x4f, 0x50,
	0x52, 0x43, 0x5f, 0x49, 0x64, 0x45, 0x72, 0x72, 0x10, 0x03, 0x12, 0x0e, 0x0a, 0x0a, 0x4f, 0x50,
	0x52, 0x43, 0x5f, 0x44, 0x62, 0x45, 0x72, 0x72, 0x10, 0x04, 0x12, 0x10, 0x0a, 0x0c, 0x4f, 0x50,
	0x52, 0x43, 0x5f, 0x42, 0x61, 0x67, 0x46, 0x75, 0x6c, 0x6c, 0x10, 0x05, 0x12, 0x12, 0x0a, 0x0e,
	0x4f, 0x50, 0x52, 0x43, 0x5f, 0x4e, 0x6f, 0x74, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x10, 0x06,
	0x2a, 0x88, 0x01, 0x0a, 0x09, 0x53, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x49, 0x44, 0x12, 0x13,
	0x0a, 0x0f, 0x50, 0x41, 0x43, 0x4b, 0x45, 0x54, 0x5f, 0x42, 0x41, 0x47, 0x5f, 0x5a, 0x45, 0x52,
	0x4f, 0x10, 0x00, 0x12, 0x18, 0x0a, 0x13, 0x50, 0x41, 0x43, 0x4b, 0x45, 0x54, 0x5f, 0x41, 0x4c,
	0x4c, 0x5f, 0x42, 0x41, 0x47, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x10, 0xe2, 0x13, 0x12, 0x17, 0x0a,
	0x12, 0x50, 0x41, 0x43, 0x4b, 0x45, 0x54, 0x5f, 0x41, 0x4c, 0x4c, 0x5f, 0x42, 0x41, 0x47, 0x5f,
	0x55, 0x53, 0x45, 0x10, 0xe3, 0x13, 0x12, 0x1a, 0x0a, 0x15, 0x50, 0x41, 0x43, 0x4b, 0x45, 0x54,
	0x5f, 0x53, 0x43, 0x5f, 0x53, 0x59, 0x4e, 0x43, 0x42, 0x41, 0x47, 0x44, 0x41, 0x54, 0x41, 0x10,
	0xe4, 0x13, 0x12, 0x17, 0x0a, 0x12, 0x50, 0x41, 0x43, 0x4b, 0x45, 0x54, 0x5f, 0x41, 0x4c, 0x4c,
	0x5f, 0x42, 0x41, 0x47, 0x5f, 0x45, 0x4e, 0x44, 0x10, 0xf5, 0x13, 0x42, 0x07, 0x5a, 0x05, 0x2e,
	0x3b, 0x62, 0x61, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_bag_proto_rawDescOnce sync.Once
	file_bag_proto_rawDescData = file_bag_proto_rawDesc
)

func file_bag_proto_rawDescGZIP() []byte {
	file_bag_proto_rawDescOnce.Do(func() {
		file_bag_proto_rawDescData = protoimpl.X.CompressGZIP(file_bag_proto_rawDescData)
	})
	return file_bag_proto_rawDescData
}

var file_bag_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_bag_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_bag_proto_goTypes = []interface{}{
	(OpResultCode)(0),     // 0: bag.OpResultCode
	(SPacketID)(0),        // 1: bag.SPacketID
	(*ItemInfo)(nil),      // 2: bag.ItemInfo
	(*CSBagInfo)(nil),     // 3: bag.CSBagInfo
	(*SCBagInfo)(nil),     // 4: bag.SCBagInfo
	(*CSUpBagInfo)(nil),   // 5: bag.CSUpBagInfo
	(*SCUpBagInfo)(nil),   // 6: bag.SCUpBagInfo
	(*SCSyncBagData)(nil), // 7: bag.SCSyncBagData
}
var file_bag_proto_depIdxs = []int32{
	0, // 0: bag.SCBagInfo.RetCode:type_name -> bag.OpResultCode
	2, // 1: bag.SCBagInfo.Infos:type_name -> bag.ItemInfo
	0, // 2: bag.SCUpBagInfo.RetCode:type_name -> bag.OpResultCode
	2, // 3: bag.SCUpBagInfo.Infos:type_name -> bag.ItemInfo
	2, // 4: bag.SCSyncBagData.Infos:type_name -> bag.ItemInfo
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_bag_proto_init() }
func file_bag_proto_init() {
	if File_bag_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_bag_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ItemInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_bag_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CSBagInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_bag_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SCBagInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_bag_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CSUpBagInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_bag_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SCUpBagInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_bag_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SCSyncBagData); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_bag_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_bag_proto_goTypes,
		DependencyIndexes: file_bag_proto_depIdxs,
		EnumInfos:         file_bag_proto_enumTypes,
		MessageInfos:      file_bag_proto_msgTypes,
	}.Build()
	File_bag_proto = out.File
	file_bag_proto_rawDesc = nil
	file_bag_proto_goTypes = nil
	file_bag_proto_depIdxs = nil
}
