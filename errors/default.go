package errors

const (
	E_SUCCESS = 0

	E_BadRequest      = 400 // 错误请求
	E_UnAuthorized    = 401 // 未授权（如未登录）
	E_Forbidden       = 403 // 拒绝访问
	E_NotFound        = 404
	E_TimeOut         = 408 // 请求超时
	E_TooManyRequests = 429 // 请求次数过多（限流）

	E_InternalServer     = 500 // 服务器内部错误
	E_ServiceUnavailable = 503 // 服务暂时不可用（如服务器过载或维护）
	E_GatewayTimeOut     = 504 // 网关超时
)

var (
	SUCCESS = add(E_SUCCESS, "success")

	ErrBadRequest      = add(E_BadRequest, "Bad request")
	ErrUnAuthorized    = add(E_UnAuthorized, "Unauthorized")
	ErrForbidden       = add(E_Forbidden, "Forbidden")
	ErrNotFound        = add(E_NotFound, "Not found")
	ErrTimeOut         = add(E_TimeOut, "Request timeout")
	ErrTooManyRequests = add(E_TooManyRequests, "Too many requests")

	ErrInternalServer     = add(E_InternalServer, "Internal server error")
	ErrServiceUnavailable = add(E_ServiceUnavailable, "Service unavailable")
	ErrGatewayTimeOut     = add(E_GatewayTimeOut, "Gateway timeout")
)
