package middleware

const (
	middlewareTypeState middlewareType = iota
	middlewareTypeIncomingLog
	middlewareTypeAdblock
	middlewareTypeRateLimiter
	middlewareTypeHeaders
	middlewareTypeReferer
	middlewareTypeSessions
	middlewareTypeProxyRequest
)
