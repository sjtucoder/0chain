package faucetsc

import (
	benchmark "0chain.net/smartcontract/benchmark"
	"0chain.net/smartcontract/rest"
)

func BenchmarkRestTests(
	data benchmark.BenchData, _ benchmark.SignatureScheme,
) benchmark.TestSuite {
	rh := rest.NewTestRestHandler()
	frh := NewFaucetscRestHandler(rh)
	return benchmark.GetRestTests(
		[]benchmark.TestParameters{
			{
				FuncName: "personalPeriodicLimit",
				Params: map[string]string{
					"client_id": data.Clients[0],
				},
				Endpoint: frh.getGlobalPeriodicLimit,
			},
			{
				FuncName: "globalPeriodicLimit",
				Endpoint: frh.getGlobalPeriodicLimit,
			},
			{
				FuncName: "pourAmount",
				Endpoint: frh.getPourAmount,
			},
			{
				FuncName: "getConfig",
				Endpoint: frh.getConfig,
			},
		},
		ADDRESS,
		frh,
	)
}
