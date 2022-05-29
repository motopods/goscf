package goscf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/tencentyun/scf-go-lib/events"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type (
	HandlerFunc func(ctx context.Context) error

	MiddlewareFunc func(next HandlerFunc) HandlerFunc
)

func apiWrapper(handler interface{}, middleware ...MiddlewareFunc) interface{} {
	handlerValue := reflect.ValueOf(handler)
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		panic(fmt.Errorf("handler kind %s is not %s", handlerType.Kind(), reflect.Func))
	}
	return func(ctx context.Context, req events.APIGatewayRequest) (events.APIGatewayResponse, error) {
		res := events.APIGatewayResponse{
			StatusCode:      http.StatusOK,
			Headers:         map[string]string{},
			IsBase64Encoded: false,
		}
		if handler == nil {
			return res, errors.New("handler is nil")
		}

		// 通过反射调用处理函数
		h := func(ctx context.Context) error {
			res := ResponseFromContext(ctx)

			var args []reflect.Value
			args = append(args, reflect.ValueOf(ctx))

			if handlerType.NumIn() == 2 {
				eventType := handlerType.In(handlerType.NumIn() - 1)
				event := reflect.New(eventType)

				if req.Method == http.MethodGet {
					err := bindData(event.Interface(), req.QueryString, "json")
					if err != nil {
						return err
					}
				} else {
					data := []byte(req.Body)
					// 转换FORM数据
					cType := Header(req.Headers).Get("content-type")
					if cType == "" {
						cType = http.DetectContentType(data)
					}
					if strings.Contains(cType, "form") {
						params, err := url.ParseQuery(req.Body)
						if err != nil {
							return err
						}
						m := map[string]interface{}{}
						for k, v := range params {
							if len(v) == 1 {
								m[k] = v[0]
							} else {
								m[k] = "[" + strings.Join(v, " ") + "]"
							}
						}
						data, err = json.Marshal(m)
						if err != nil {
							return err
						}
					}
					// 绑定JSON数据
					if err := json.Unmarshal(data, event.Interface()); err != nil {
						return err
					}
				}

				args = append(args, event.Elem())
			}

			response := handlerValue.Call(args)

			if len(response) > 0 {
				// 检查是否有错误
				if len(response) == 2 && !response[1].IsNil() {
					return response[1].Interface().(error)
				} else {
					resData := response[0]
					if resData.Kind() == reflect.Interface {
						resData = resData.Elem()
					}
					if resData.Kind() == reflect.String {
						// 如果是字符串则不转换
						res.Headers["Content-Type"] = "application/x-javascript;charset=utf-8"
						res.Body = resData.String()
					} else {
						// 如果是struct则转为json
						res.Headers["Content-Type"] = "application/json;charset=utf-8"
						var data interface{}
						if resData.IsValid() {
							data = resData.Interface()
						}
						payload, err := json.Marshal(map[string]interface{}{
							"data": data,
						})
						if err != nil {
							return err
						}
						res.Body = string(payload)
					}
				}
			}

			return nil
		}

		// 应用中间件
		for i := len(middleware) - 1; i >= 0; i-- {
			h = middleware[i](h)
		}

		// 执行函数
		ctx = ContextWithRequest(ctx, &req)
		ctx = ContextWithResponse(ctx, &res)
		ctx = ContextWithIp(ctx, req.Context.SourceIP)
		if err := h(ctx); err != nil {
			payload, _ := json.Marshal(map[string]interface{}{
				"error": err.Error(),
			})
			res.Body = string(payload)
			return res, nil
		}

		return res, nil
	}
}

func APIGateway(handler interface{}, middleware ...MiddlewareFunc) {
	wrapperHandler := apiWrapper(handler, middleware...)
	cloudfunction.Start(wrapperHandler)
}
