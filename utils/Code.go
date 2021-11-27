package utils

type ResCode int64

const (
	CodeSuccess ResCode = 1000 + iota
	CodeServerErr
	CodeErr
)

var codeMsgCode = map[ResCode]string{
	CodeSuccess:   "success",
	CodeServerErr: "服务器错误",
	CodeErr:       "错误",
}

func (r ResCode) Msg() string {
	codeMsg, ok := codeMsgCode[r]
	if !ok {
		codeMsg = codeMsgCode[CodeServerErr]
	}
	return codeMsg
}
