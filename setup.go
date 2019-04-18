package rewriteng

import (
	"fmt"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/mholt/caddy"
	"github.com/miekg/dns"
)

const (
	answerRule     = "answer"
	additionalRule = "additional"
	authorityRule  = "authority"
)

func init() {
	caddy.RegisterPlugin("rewriteng", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) (err error) {
	var rules []*nameRule

	if rules, err = parse(c); err != nil {
		return plugin.Error("rewriteng", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return RewriteNG{Next: next, Rules: rules}
	})

	return
}

func parse(c *caddy.Controller) (rules []*nameRule, err error) {
	for c.Next() {
		var af, ok bool
		var rule nameRule
		var r, innerRule Rule
		var from, to string

		// CLASS
		if !c.NextArg() {
			return rules, c.ArgErr()
		}
		if rule.RRClass, ok = dns.StringToClass[c.Val()]; !ok {
			return rules, c.Errf("invalid query class %s", c.Val())
		}

		// RR TYPE
		if !c.NextArg() {
			return rules, c.ArgErr()
		}
		if rule.RRType, ok = dns.StringToType[c.Val()]; !ok {
			return rules, c.Errf("invalid RR class %s", c.Val())
		}

		// RULE TYPE
		if !c.NextArg() {
			return rules, c.ArgErr()
		}
		rule.RuleType = c.Val()
		if err = checkRuleType(rule.RuleType); err != nil {
			return
		}

		// STRING
		if !c.NextArg() {
			return rules, c.ArgErr()
		}
		from = c.Val()

		// STRING
		if !c.NextArg() {
			return rules, c.ArgErr()
		}
		to = c.Val()

		if innerRule, err = newRule(rule.RuleType, from, to); err != nil {
			return
		}
		rule.Rule = innerRule

		for c.NextBlock() {
			v := c.Val()
			args := c.RemainingArgs()
			if len(args) < 3 {
				return rules, c.ArgErr()
			}

			switch v {
			case answerRule:
				if !af {
					af = true
				}
				if r, err = newRule(args...); err != nil {
					return
				}
				rule.Answers = append(rule.Answers, r)
			case additionalRule:
				if r, err = newRule(args...); err != nil {
					return
				}
				rule.Additional = append(rule.Additional, r)
			case authorityRule:
				if r, err = newRule(args...); err != nil {
					return
				}
				rule.Authority = append(rule.Authority, r)
			default:
				return rules, c.Errf("Only answer, additional, authority supported, received: %s", v)
			}
		}

		if !af {
			return rules, c.Errf("atleast one answer is required")
		}
		rules = append(rules, &rule)
	}
	return
}

func checkRuleType(r string) (err error) {
	switch r {
	case ExactMatch:
	case PrefixMatch:
	case SuffixMatch:
	case SubstringMatch:
	case RegexMatch:
	default:
		err = fmt.Errorf("invalid rule type %q", r)
	}
	return
}
