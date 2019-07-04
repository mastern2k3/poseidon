package storage

import (
	"context"
	"encoding/json"

	"github.com/heroiclabs/nakama/api"
	"github.com/heroiclabs/nakama/runtime"
)

type KeysetCollectionAccessor struct {
	CollectionID string
	ModelFactory func() interface{}
}

type KeyedValue struct {
	Key   string
	Value interface{}
}

func (acc *KeysetCollectionAccessor) Get(ctx context.Context, nk runtime.NakamaModule, userID string) ([]KeyedValue, error) {

	var allObjs []*api.StorageObject
	var lastCursor string

	for {
		objs, newCur, err := nk.StorageList(ctx, userID, acc.CollectionID, 100, lastCursor)
		if err != nil {
			return nil, err
		}

		if newCur == "" || newCur == lastCursor {
			break
		}

		lastCursor = newCur
		allObjs = append(allObjs, objs...)
	}

	var res []KeyedValue

	for _, read := range allObjs {

		model := acc.ModelFactory()

		err := json.Unmarshal([]byte(read.GetValue()), model)

		if err != nil {
			return nil, err
		}

		res = append(res, KeyedValue{
			Key:   read.Key,
			Value: model,
		})
	}

	return res, nil
}

func (acc *KeysetCollectionAccessor) Save(ctx context.Context, nk runtime.NakamaModule, userID string, kv KeyedValue) error {

	bytes, err := json.Marshal(kv.Value)

	if err != nil {
		return err
	}

	_, err = nk.StorageWrite(ctx, []*runtime.StorageWrite{
		&runtime.StorageWrite{
			UserID:     userID,
			Collection: acc.CollectionID,
			Key:        kv.Key,
			Value:      string(bytes),
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (acc *KeysetCollectionAccessor) SaveList(ctx context.Context, nk runtime.NakamaModule, userID string, data []KeyedValue) error {

	writes := []*runtime.StorageWrite{}

	for _, d := range data {
		bytes, err := json.Marshal(d.Value)
		if err != nil {
			return err
		}
		writes = append(writes, &runtime.StorageWrite{
			UserID:     userID,
			Collection: acc.CollectionID,
			Key:        d.Key,
			Value:      string(bytes),
		})
	}

	_, err := nk.StorageWrite(ctx, writes)

	if err != nil {
		return err
	}

	return nil
}

func (acc *KeysetCollectionAccessor) GetList(ctx context.Context, nk runtime.NakamaModule, userIDs []string) (map[string][]KeyedValue, error) {

	resp := map[string][]KeyedValue{}

	for _, userID := range userIDs {

		vals, err := acc.Get(ctx, nk, userID)
		if err != nil {
			return nil, err
		}

		resp[userID] = vals
	}

	return resp, nil
}

func (acc *KeysetCollectionAccessor) Delete(ctx context.Context, nk runtime.NakamaModule, key string, userID string) error {

	return nk.StorageDelete(ctx, []*runtime.StorageDelete{
		&runtime.StorageDelete{
			Collection: acc.CollectionID,
			Key:        key,
			UserID:     userID,
		},
	})
}
