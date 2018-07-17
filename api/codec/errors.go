package codec

import "github.com/zhsyourai/URCF-engine/api"

func GetErrorCode(err error) int {
	switch err.(type) {
	case *api.MethodNotFoundError:
		return -32601
	case *api.InvalidRequestError:
		return -32600
	case *api.InvalidMessageError:
		return -32700
	case *api.InvalidParamsError:
		return -32602
	case *api.CallbackError:
		return -32000
	case *api.ShutdownError:
		return -32000
	}
	return 0
}
