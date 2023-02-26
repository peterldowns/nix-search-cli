package nixsearch

import (
	"context"
)

type Client interface {
	Search(ctx context.Context, query Query) ([]Package, error)
}
