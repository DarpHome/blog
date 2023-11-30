package globals

import goflaker "github.com/MCausc78/goflaker"

const (
	SnowflakesEpoch uint64 = 1699567200000
)

var Snowflakes *goflaker.DefaultSnowflakeGenerator
