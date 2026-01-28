package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/SadikSunbul/gopilot"
	"github.com/SadikSunbul/gopilot/provider"
)

/*
Senaryo: Plan step retry (flaky function)

Amaç:
- LLM bir plan döndürür.
- İlk step "flaky" ilk denemede hata verir, ikinci denemede başarılı olur.
- Plan step'inde retries=1 olduğu için orkestratör tekrar dener ve plan tamamlanır.

Çalıştırma:
  go run ./examples/retry_plan
*/

type planProvider struct{}

func (p *planProvider) Generate(ctx context.Context, prompt string) (*provider.Response, error) {
	_ = ctx
	_ = prompt
	return &provider.Response{
		Type: "plan",
		Steps: []provider.Step{
			{Function: "flaky", Retries: 1, StoreAs: "flaky", MapOutput: boolPtr(false)},
			{Function: "ok", Params: map[string]any{"value": "$flaky.value"}},
		},
	}, nil
}

func (p *planProvider) Close() error { return nil }

func boolPtr(v bool) *bool { return &v }

type FlakyParams struct{}
type FlakyOut struct {
	Value string `json:"value" description:"Output value produced by flaky step" example:"ok"`
}

type OKParams struct {
	Value string `json:"value" description:"Value passed from previous step" required:"true" example:"ok"`
}
type OKOut struct {
	Done bool `json:"done" description:"Indicates plan completed successfully" example:"true"`
}

func buildRetryPlanGopilot(counter *int32) (*gopilot.Gopilot, error) {
	g, err := gopilot.New(&planProvider{})
	if err != nil {
		return nil, err
	}

	flaky := func(ctx context.Context, _ FlakyParams) (FlakyOut, error) {
		_ = ctx
		if atomic.AddInt32(counter, 1) == 1 {
			return FlakyOut{}, errors.New("temporary error")
		}
		return FlakyOut{Value: "ok"}, nil
	}

	okFn := func(ctx context.Context, p OKParams) (OKOut, error) {
		_ = ctx
		return OKOut{Done: p.Value == "ok"}, nil
	}

	if err := g.Register(gopilot.NewFunction[FlakyParams, FlakyOut]("flaky", "fails once then succeeds", flaky)); err != nil {
		return nil, err
	}
	if err := g.Register(gopilot.NewFunction[OKParams, OKOut]("ok", "final step", okFn)); err != nil {
		return nil, err
	}

	return g, nil
}

func main() {
	var counter int32
	g, err := buildRetryPlanGopilot(&counter)
	if err != nil {
		panic(err)
	}

	out, err := g.GenerateAndExecuteSmart(context.Background(), "run")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Result: %+v (attempts=%d)\n", out, atomic.LoadInt32(&counter))
}
