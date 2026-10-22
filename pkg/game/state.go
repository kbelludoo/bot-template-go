package game

import (
	"bytes"
	"encoding/json"
)

// UnmarshalJSON preserva campos desconhecidos e inteiros grandes para futuros
// envelopes de save. Nunca derive um save apenas dos campos tipados desta struct.
func (s *PokeTrueState) UnmarshalJSON(data []byte) error {
	type stateFields PokeTrueState
	var fields stateFields
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	var raw map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return err
	}
	fields.RawEstado = raw
	*s = PokeTrueState(fields)
	return nil
}
