package middleware

import (
	"context"
	"github.com/motopods/goscf"
	"net/http"
	"strconv"
	"strings"
)

type (
	CORSConfig struct {
		// 允许的源，一般为源站域名，*代表所有
		// 可选. 默认值 []string{"*"}.
		AllowOrigins []string

		// 允许的方法，用在preflight请求，即OPTIONS请求。
		// 可选. 默认值 DefaultCORSConfig.AllowMethods.
		AllowMethods []string

		// 允许的头
		// 可选. 默认值 []string{}.
		AllowHeaders []string

		// 允许携带cookie
		// 可选. 默认值 false.
		AllowCredentials bool

		// 客户端允许的访问的白名单响应头
		// 可选. 默认值 []string{}.
		ExposeHeaders []string

		// 过期时间
		MaxAge int
	}
)

var (
	// DefaultCORSConfig 默认的CORS配置，允许所有
	DefaultCORSConfig = CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut,
			http.MethodPatch, http.MethodPost, http.MethodDelete},
	}
)

func CORS() goscf.MiddlewareFunc {
	return CORSWithConfig(&DefaultCORSConfig)
}

func CORSWithConfig(config *CORSConfig) goscf.MiddlewareFunc {
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultCORSConfig.AllowOrigins
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig.AllowMethods
	}

	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := strconv.Itoa(config.MaxAge)

	return func(next goscf.HandlerFunc) goscf.HandlerFunc {
		return func(ctx context.Context) error {
			req := goscf.RequestFromContext(ctx)
			res := goscf.ResponseFromContext(ctx)
			reqHeader := goscf.Header(req.Headers)
			resHeader := goscf.Header(goscf.ResponseFromContext(ctx).Headers)

			origin := reqHeader.Get("Origin")
			allowOrigin := ""

			// 检查允许的域名
			for _, o := range config.AllowOrigins {
				if o == "*" && config.AllowCredentials {
					allowOrigin = origin
					break
				}
				if o == "*" || o == origin {
					allowOrigin = o
					break
				}
				if matchSubDomain(origin, o) {
					allowOrigin = origin
					break
				}
			}

			// 简单请求
			if req.Method != http.MethodOptions {
				resHeader.Add("Vary", "Origin")
				resHeader.Set("Access-Control-Allow-Origin", allowOrigin)
				if config.AllowCredentials {
					resHeader.Set("Access-Control-Allow-Credentials", "true")
				}
				if exposeHeaders != "" {
					resHeader.Set("Access-Control-Expose-Headers", exposeHeaders)
				}
				return next(ctx)
			}

			// 预先请求
			resHeader.Add("Vary", "Origin")
			resHeader.Add("Vary", "Access-Control-Request-Method")
			resHeader.Add("Vary", "Access-Control-Request-Headers")
			resHeader.Set("Access-Control-Allow-Origin", allowOrigin)
			resHeader.Set("Access-Control-Allow-Methods", allowMethods)
			if config.AllowCredentials {
				resHeader.Set("Access-Control-Allow-Credentials", "true")
			}
			if allowHeaders != "" {
				resHeader.Set("Access-Control-Allow-Headers", allowHeaders)
			} else {
				h := reqHeader.Get("Access-Control-Request-Headers")
				if h != "" {
					resHeader.Set("Access-Control-Allow-Headers", h)
				}
			}
			if config.MaxAge > 0 {
				resHeader.Set("Access-Control-Max-Age", maxAge)
			}
			res.StatusCode = http.StatusNoContent
			return nil
		}
	}
}
