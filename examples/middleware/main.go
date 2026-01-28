package main

import (
	"context"
	"fmt"

	"github.com/SadikSunbul/gopilot"
)

/*
Senaryo: Middleware ile parametre/enstrümantasyon

Amaç:
- Her fonksiyon çalıştırılmadan önce parametrelere request_id enjekte et.
- Böylece tüm fonksiyonlar ortak bir tracing bilgisi alabilir.

Çalıştırma:
  go run ./examples/middleware
*/

type EchoParams struct {
	Message   string `json:"message" description:"Message to echo back" required:"true" example:"hello"`
	RequestID string `json:"request_id,omitempty" description:"Request correlation id (injected by middleware)" example:"req-123"`
}

type EchoResp struct {
	Out       string `json:"out" description:"Echo output" example:"echo: hello"`
	RequestID string `json:"request_id" description:"Request correlation id that was used" example:"req-123"`
}

func Echo(ctx context.Context, p EchoParams) (EchoResp, error) {
	_ = ctx
	return EchoResp{Out: "echo: " + p.Message, RequestID: p.RequestID}, nil
}

func buildMiddlewareGopilot() (*gopilot.Gopilot, error) {
	injectRequestID := func(next gopilot.ExecuteFunc) gopilot.ExecuteFunc {
		return func(name string, params map[string]any) (any, error) {
			if params == nil {
				params = map[string]any{}
			}
			if _, ok := params["request_id"]; !ok {
				params["request_id"] = "req-123"
			}
			return next(name, params)
		}
	}

	g, err := gopilot.New(&noopProvider{}, gopilot.WithMiddleware(injectRequestID))
	if err != nil {
		return nil, err
	}

	if err := g.Register(gopilot.NewFunction[EchoParams, EchoResp](
		"echo",
		"Echo message",
		Echo,
	)); err != nil {
		return nil, err
	}

	return g, nil
}

func main() {
	g, err := buildMiddlewareGopilot()
	if err != nil {
		panic(err)
	}

	out, err := g.Execute(context.Background(), "echo", map[string]any{"message": "hi"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result: %+v\n", out)
}
