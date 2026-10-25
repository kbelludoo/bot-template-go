package protocol

import (
	"encoding/json"
	"fmt"
)

// FormatTRPCBatchInput envelopa o payload no formato tRPC v10 Batch `{"0":{"json":...}}` (usado com `?batch=1`)
func FormatTRPCBatchInput(input any) ([]byte, error) {
	envelope := map[string]any{
		"0": map[string]any{
			"json": input,
		},
	}
	return json.Marshal(envelope)
}

// UnwrapTRPCBatchResponse desenvelopa a resposta tRPC v10 `[{"result":{"data":{"json":...}}}]` para a struct alvo
func UnwrapTRPCBatchResponse(body []byte, target any) error {
	var batch []struct {
		Result *struct {
			Data struct {
				JSON json.RawMessage `json:"json"`
			} `json:"data"`
		} `json:"result"`
		Error *struct {
			JSON struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
			} `json:"json"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &batch); err != nil {
		return fmt.Errorf("resposta tRPC batch malformada: %w", err)
	}
	if len(batch) == 0 {
		return fmt.Errorf("resposta tRPC batch vazia")
	}
	if batch[0].Error != nil && batch[0].Error.JSON.Message != "" {
		return fmt.Errorf("erro tRPC (%d): %s", batch[0].Error.JSON.Code, batch[0].Error.JSON.Message)
	}
	if batch[0].Result == nil || len(batch[0].Result.Data.JSON) == 0 {
		return nil
	}
	if target == nil {
		return nil
	}
	return json.Unmarshal(batch[0].Result.Data.JSON, target)
}
