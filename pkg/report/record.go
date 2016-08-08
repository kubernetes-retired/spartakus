package report

import "encoding/json"

type Record struct {
	Metadata map[string]string
	Payload  interface{}
}

func (r *Record) UnmarshalJSON(d []byte) error {
	var jr jsonRecord
	if err := json.Unmarshal(d, &jr); err != nil {
		return err
	}

	r.Metadata = jr.Metadata
	r.Payload = jr.Payload

	return nil
}

func (r *Record) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(r.Payload)
	if err != nil {
		return nil, err
	}
	msg := json.RawMessage(b)

	jr := jsonRecord{
		Metadata: r.Metadata,
		Payload:  &msg,
	}

	return json.Marshal(&jr)
}

type jsonRecord struct {
	Metadata map[string]string `json:"metadata"`
	Payload  *json.RawMessage  `json:"payload"`
}
