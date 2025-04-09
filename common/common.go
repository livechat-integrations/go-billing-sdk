package common

import "context"

type TokenFn func(ctx context.Context) (string, error)
