package ledger

import "context"

// ListParties returns all parties on the ledger.
func (c *HTTPLedgerClient) ListParties(ctx context.Context) ([]Party, error) {
	data, err := c.doRequest(ctx, "GET", "/v2/parties", nil)
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]Party](data)
}

// AllocateParty creates a new party on the ledger.
func (c *HTTPLedgerClient) AllocateParty(ctx context.Context, name, displayName string) (*Party, error) {
	body, err := encodeJSON(map[string]string{
		"party":        name,
		"display_name": displayName,
	})
	if err != nil {
		return nil, err
	}
	data, err := c.doRequest(ctx, "POST", "/v2/parties/allocate", body)
	if err != nil {
		return nil, err
	}
	return decodeJSONPtr[Party](data)
}

func decodeJSONPtr[T any](data []byte) (*T, error) {
	v, err := decodeJSON[T](data)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
