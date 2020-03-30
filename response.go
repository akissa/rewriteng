package rewriteng

import (
	"net"
	"strconv"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

// ResponseRewriter rewrites answers, additional and authority sections
type ResponseRewriter struct {
	dns.ResponseWriter
	originalQuestion dns.Question
	Rules            []*nameRule
	recursionDesired bool
}

// NewResponseRewriter returns a pointer to a new ResponseRewriter.
func NewResponseRewriter(w dns.ResponseWriter, r *dns.Msg) *ResponseRewriter {
	return &ResponseRewriter{
		ResponseWriter:   w,
		originalQuestion: r.Question[0],
		recursionDesired: r.RecursionDesired,
	}
}

// WriteMsg records the status code and calls the underlying ResponseWriter's WriteMsg method.
func (r *ResponseRewriter) WriteMsg(res *dns.Msg) error {
	res.Question[0] = r.originalQuestion
	res.RecursionDesired = r.recursionDesired
	// Answers
	r.rewriteAnswers(res)
	// Authority
	r.rewriteAuthority(res)
	// Additional
	r.rewriteAdditional(res)

	return r.ResponseWriter.WriteMsg(res)
}

func (r *ResponseRewriter) rewriteAnswers(res *dns.Msg) {

	log.Debugf("In ResponseWriter %v", res)

	for _, rr := range res.Answer {
		var nameWriten bool
		var name = rr.Header().Name
		for _, rule := range r.Rules {
			if !nameWriten {
				if s := rule.Sub(name, answerRule, namePart); s != "" {
					rr.Header().Name = s
					nameWriten = true
				}
			}
			// Rewrite Data
			log.Debugf("calling rewriteAnswers -> rewriteDataParts for rr %v rule %v answerRule %s ", rr, rule, answerRule)
			rewriteDataParts(rr, rule, answerRule)
		}
	}
}

func (r *ResponseRewriter) rewriteAuthority(res *dns.Msg) {
	for _, rr := range res.Ns {
		var nameWriten bool
		var name = rr.Header().Name
		for _, rule := range r.Rules {
			if !nameWriten {
				if s := rule.Sub(name, authorityRule, namePart); s != "" {
					rr.Header().Name = s
					nameWriten = true
				}
			}
			// Rewrite Data
			log.Debugf("calling rewriteAuthority -> rewriteDataParts for rr %v rule %v answerRule %s ", rr, rule, authorityRule)
			rewriteDataParts(rr, rule, authorityRule)
		}
	}
}

func (r *ResponseRewriter) rewriteAdditional(res *dns.Msg) {
	for _, rr := range res.Extra {
		var nameWriten bool
		var name = rr.Header().Name
		for _, rule := range r.Rules {
			if !nameWriten {
				if s := rule.Sub(name, additionalRule, namePart); s != "" {
					rr.Header().Name = s
					nameWriten = true
				}
			}
			// Rewrite Data
			log.Debugf("calling rewriteAdditional-> rewriteDataParts for rr %v rule %v answerRule %s ", rr, rule, additionalRule)
			rewriteDataParts(rr, rule, additionalRule)
		}
	}
}

func rewriteDataParts(rr dns.RR, rule *nameRule, ruleType string) {
	switch t := rr.(type) {
	case *dns.CNAME:
		if s := rule.Sub(t.Target, ruleType, dataPart); s != "" {
			t.Target = plugin.Name(s).Normalize()
		}
	case *dns.TXT:
		tmpRecs := t.Txt[:0]
		for _, txtr := range t.Txt {
			if s := rule.Sub(txtr, ruleType, dataPart); s != "" {
				tmpRecs = append(tmpRecs, s)
			} else {
				tmpRecs = append(tmpRecs, txtr)
			}
		}
		t.Txt = tmpRecs
	case *dns.A:
		if s := rule.Sub(t.A.String(), ruleType, dataPart); s != "" {
			if ip := net.ParseIP(s); ip != nil && strings.Contains(s, ".") {
				t.A = ip
			}
		}
	case *dns.AAAA:
		if s := rule.Sub(t.AAAA.String(), ruleType, dataPart); s != "" {
			if ip := net.ParseIP(s); ip != nil && strings.Contains(s, ":") {
				t.AAAA = ip
			}
		}
	case *dns.NS:
		if s := rule.Sub(t.Ns, ruleType, dataPart); s != "" {
			t.Ns = plugin.Name(s).Normalize()
		}
	case *dns.SOA:
		if s := rule.Sub(t.Ns, ruleType, dataPart); s != "" {
			t.Ns = plugin.Name(s).Normalize()
		}
		if s := rule.Sub(t.Mbox, ruleType, dataPart); s != "" {
			t.Mbox = plugin.Name(s).Normalize()
		}
		if s := rule.Sub(strconv.FormatInt(int64(t.Minttl), 10), ruleType, dataPart); s != "" {
			if i, err := strconv.ParseUint(s, 10, 32); err == nil {
				t.Minttl = uint32(i)
			}
		}
	case *dns.SRV:
		if s := rule.Sub(t.Target, ruleType, dataPart); s != "" {
			t.Target = plugin.Name(s).Normalize()
		}
	case *dns.PTR:
		log.Debugf("Rewrite PTR  rr \"%s\"", rr)
		if s := rule.Sub(t.Ptr, ruleType, dataPart); s != "" {
			log.Debugf("Replacement %s", s)
			t.Ptr = plugin.Name(s).Normalize()
			log.Debugf("Rewrite PTR  t.ptr %s", t.Ptr)
		}
	}
}
