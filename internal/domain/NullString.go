package domain

import (
	"database/sql"
	"encoding/json"
)

// NullString is sql.NullString with clean JSON. By default sql.NullString
// serializes as the struct it is — {"String":"...","Valid":true} — which would
// leak a database-layer shape onto our public API. NullString overrides that to
// emit the bare string when present, or null when not.
type NullString struct {
	sql.NullString
}

// MarshalJSON emits the bare string, or null when the value is absent.
func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.String)
}

// UnmarshalJSON is the inverse — accepts the bare string or null — so the type
// round-trips cleanly if it's ever decoded from a request body too.
func (n *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid, n.String = false, ""
		return nil
	}
	if err := json.Unmarshal(data, &n.String); err != nil {
		return err
	}
	n.Valid = true
	return nil
}
