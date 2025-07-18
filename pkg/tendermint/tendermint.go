package tendermint

import (
	"context"
	"fmt"
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type RPC struct {
	Client *http.Client
	URL    string
	Logger zerolog.Logger
	Tracer trace.Tracer

	LastQueryHeight map[string]int64
	Mutex           sync.Mutex
}

func NewRPC(chain config.Chain, logger zerolog.Logger, tracer trace.Tracer) *RPC {
	return &RPC{
		Client:          http.NewClient(logger, chain.Name, tracer),
		URL:             chain.LCDEndpoint,
		Logger:          logger.With().Str("component", "rpc").Logger(),
		LastQueryHeight: make(map[string]int64),
		Tracer:          tracer,
	}
}

func (rpc *RPC) GetWalletBalances(address string, ctx context.Context) (*types.BalanceResponse, types.QueryInfo, error) {
	lastHeight, _ := rpc.LastQueryHeight[address]

	url := fmt.Sprintf(
		"%s/ethereum/bank/v1beta1/balances/%s",
		rpc.URL,
		address,
	)

	var response *types.BalanceResponse
	queryInfo, header, err := rpc.Client.Get(url, &response, types.HTTPPredicateCheckHeightAfter(lastHeight), ctx)
	if err != nil {
		return nil, queryInfo, err
	}

	newLastHeight, _ := utils.GetBlockHeightFromHeader(header)

	rpc.Mutex.Lock()
	rpc.LastQueryHeight[address] = newLastHeight
	rpc.Mutex.Unlock()

	return response, queryInfo, nil
}
