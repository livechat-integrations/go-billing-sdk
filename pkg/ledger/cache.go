package ledger

type Cache interface {
	Set(k string, v float32) bool
	Get(k string) (float32, bool)
}
