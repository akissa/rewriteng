package rewriteng

import (
	"context"
	"testing"

	"github.com/miekg/dns"
)

func msgPrinter(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	w.WriteMsg(r)
	return 0, nil
}

func TestRewriteNG(t *testing.T) {
}
