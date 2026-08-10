package job

import "github.com/google/wire"

// ProviderSet is job providers.
var ProviderSet = wire.NewSet(NewSettlementJob)
