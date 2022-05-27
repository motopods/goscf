package goscf

import (
	"context"
	"github.com/tencentyun/scf-go-lib/events"
)

func RequestFromContext(ctx context.Context) *events.APIGatewayRequest {
	return ctx.Value("request").(*events.APIGatewayRequest)
}

func ContextWithRequest(ctx context.Context, request *events.APIGatewayRequest) context.Context {
	return context.WithValue(ctx, "request", request)
}

func ResponseFromContext(ctx context.Context) *events.APIGatewayResponse {
	return ctx.Value("response").(*events.APIGatewayResponse)
}

func ContextWithResponse(ctx context.Context, response *events.APIGatewayResponse) context.Context {
	return context.WithValue(ctx, "response", response)
}

func IpFromContext(ctx context.Context) string {
	if ctx.Value("ip") != nil {
		return ctx.Value("ip").(string)
	}
	return ""
}

func ContextWithIp(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, "ip", ip)
}
