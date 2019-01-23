
# poseidon
> Extended apis and utilities on top of Nakama's api

## Features

### Storage Collection Accessors

Allows for more fluent use of Nakama's storage collections, taking care of json marshalling and unmarshalling, default values and bulk reads.

```go

import (
	"context"

	"github.com/heroiclabs/nakama/runtime"

	"github.com/mastern2k3/poseidon/runtime/storage"
)

var (
	statsAccessor = &storage.CollectionAccessor{
		CollectionID: "stats",
        KeyID:        "matchesPlayed",
        // An empty model used for deserializing
        ModelFactory: func() interface{} { return &MatchStats{} },
        // A default value for when a record does not exist
        DefaultFactory: func() interface{} {
			return &MatchStats{
				WinningStreak: 0,
				MatchesPlayed: 0,
			}
		},
	}
)

type MatchStats struct {
	MatchesPlayed uint `json:"matchesPlayed"`
	WinningStreak uint `json:"winningStreak"`
}

func GetMatchStats(ctx context.Context, nk runtime.NakamaModule, userIDs []string) (map[string]*MatchStats, error) {

	dx, err := statsAccessor.GetOrDefaultList(ctx, nk, userIDs)

	if err != nil {
		return nil, err
	}

	stats := map[string]*MatchStats{}

	for userID, d := range dx {
		stats[userID] = d.(*MatchStats)
	}

	return stats, nil
}
```

## RPC Routes

> Coming soon
