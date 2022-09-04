package trinity

import "context"

type trinity struct {
}

func New(ctx context.Context) *trinity {
	return &trinity{}
}

func (t *trinity) ServeHTTP(addr ...string) error {
	return nil
}
