package rewriteng

import (
	"context"
	"fmt"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

var (
	log = clog.NewWithPlugin("rewriteng")
)

// RewriteNG implements the rewriteng plugin
type RewriteNG struct {
	Next  plugin.Handler
	Rules []*nameRule
}

// Name implements the Handler interface.
func (h RewriteNG) Name() string {
	return "rewriteng"
}

// ServeDNS implements the plugin.Handler interface.
func (h RewriteNG) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	wr := NewResponseRewriter(w, r)
	log.Infof("plug-in calls RewriteNG")
	state := request.Request{W: w, Req: r}
	for _, rule := range h.Rules {
		log.Infof("plug-in calls RewriteNG")
		switch result := rule.Rewrite(ctx, state); result {
		case RewriteDone:
			if !validName(state.Req.Question[0].Name) {
				x := state.Req.Question[0].Name
				log.Errorf("Invalid name after rewrite: %s", x)
				state.Req.Question[0] = wr.originalQuestion
				return dns.RcodeServerFailure, fmt.Errorf("invalid name after rewrite: %s", x)
			}
			wr.Rules = append(wr.Rules, rule)
			log.Infof("Calling next plugin: %s", h.Name())
			return plugin.NextOrFailure(h.Name(), h.Next, ctx, wr, r)
		default:
		}
	}
	if len(wr.Rules) == 0 {
		return plugin.NextOrFailure(h.Name(), h.Next, ctx, w, r)
	}
	return plugin.NextOrFailure(h.Name(), h.Next, ctx, wr, r)
}
