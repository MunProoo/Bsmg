package main

const (
	ProcessSuccess        = 0
	ErrorInvalidParameter = 1
	ResultIsNull          = 2
	LoginSuccess          = 0
	LoginIDNotExist       = 3
	LoginPWMismatch       = 4
	NotLoggedIn           = 5
)

type WCResult struct {
	Result WCResultData `protobuf:"bytes,1,rep,name=Result" json:"Result"`
}

type WCResultData struct {
	ResultCode int32 `protobuf:"varint,1,opt,name=ResultCode" json:"ResultCode"`
}

////////////////////////////////////////////////////////////////////////////

type WResultTotal struct {
	Count int32 `protobuf:"varint,6,opt,name=Count" json:"Count"`
}
type WResultData struct {
	ResultCode int32 `protobuf:"varint,6,opt,name=ResultCode" json:"ResultCode"`
}

type WCCountResultData struct {
	Count int32 `protobuf:"bytes,2,rep,name=Count" json:"Count"`
}
