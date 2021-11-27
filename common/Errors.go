package common

import "errors"

var (
	ERR_LOCK_ALREADY_REQUIRED = errors.New("锁被占用")
	ERR_LOCAL_IP_NOT_FOUND    = errors.New("ip没有找到")
)
