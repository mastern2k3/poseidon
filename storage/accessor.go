package storage

import (
	"context"
	"encoding/json"

	"github.com/heroiclabs/nakama/runtime"
)

type CollectionAccessor struct {
	CollectionID   string
	KeyID          string
	ModelFactory   func() interface{}
	DefaultFactory func() interface{}
}

func (acc *CollectionAccessor) Get(ctx context.Context, nk runtime.NakamaModule, userID string) (interface{}, bool, error) {

	reads, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		&runtime.StorageRead{
			UserID:     userID,
			Collection: acc.CollectionID,
			Key:        acc.KeyID,
		},
	})

	if err != nil {
		return nil, false, err
	}

	if len(reads) > 0 {

		model := acc.ModelFactory()

		err := json.Unmarshal([]byte(reads[0].GetValue()), model)

		if err != nil {
			return nil, false, err
		}

		return model, true, nil
	}

	return nil, false, nil
}

func (acc *CollectionAccessor) GetOrDefault(ctx context.Context, nk runtime.NakamaModule, userID string) (interface{}, error) {

	d, f, err := acc.Get(ctx, nk, userID)

	if err != nil {
		return nil, err
	}

	if !f {
		return acc.DefaultFactory(), nil
	}

	return d, nil
}

func (acc *CollectionAccessor) Save(ctx context.Context, nk runtime.NakamaModule, userID string, data interface{}) error {

	bytes, err := json.Marshal(data)

	if err != nil {
		return err
	}

	_, err = nk.StorageWrite(ctx, []*runtime.StorageWrite{
		&runtime.StorageWrite{
			UserID:     userID,
			Collection: acc.CollectionID,
			Key:        acc.KeyID,
			Value:      string(bytes),
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (acc *CollectionAccessor) SaveList(ctx context.Context, nk runtime.NakamaModule, data map[string]interface{}) error {

	writes := []*runtime.StorageWrite{}

	for userID, d := range data {
		bytes, err := json.Marshal(d)
		if err != nil {
			return err
		}
		writes = append(writes, &runtime.StorageWrite{
			UserID:     userID,
			Collection: acc.CollectionID,
			Key:        acc.KeyID,
			Value:      string(bytes),
		})
	}

	_, err := nk.StorageWrite(ctx, writes)

	if err != nil {
		return err
	}

	return nil
}

func (acc *CollectionAccessor) GetList(ctx context.Context, nk runtime.NakamaModule, userIDs []string) (map[string]interface{}, error) {

	var reads []*runtime.StorageRead

	for _, userID := range userIDs {
		reads = append(reads, &runtime.StorageRead{
			UserID:     userID,
			Collection: acc.CollectionID,
			Key:        acc.KeyID,
		})
	}

	objs, err := nk.StorageRead(ctx, reads)

	if err != nil {
		return nil, err
	}

	responses := map[string]interface{}{}

	for _, obj := range objs {

		model := acc.ModelFactory()

		err := json.Unmarshal([]byte(obj.GetValue()), model)

		if err != nil {
			return nil, err
		}

		responses[obj.GetUserId()] = model
	}

	return responses, nil
}

func (acc *CollectionAccessor) GetOrDefaultList(ctx context.Context, nk runtime.NakamaModule, userIDs []string) (map[string]interface{}, error) {

	res, err := acc.GetList(ctx, nk, userIDs)

	if err != nil {
		return nil, err
	}

	for _, userID := range userIDs {
		_, has := res[userID]
		if !has {
			res[userID] = acc.DefaultFactory()
		}
	}

	return res, nil
}
