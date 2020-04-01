package rewriteng

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

const (
	// ExactMatch matches only on exact match of the part
	ExactMatch = "exact"
	// PrefixMatch matches when the part begins with the matching string
	PrefixMatch = "prefix"
	// SuffixMatch matches when the part ends with the matching string
	SuffixMatch = "suffix"
	// SubstringMatch matches on partial match of the part
	SubstringMatch = "substring"
	// RegexMatch matches when the part matches a regular expression and the regex is used in the rewrite
	RegexMatch = "regex"
	// FullRegexMatch matchs when the part matches a regular expression and the regex is not used in the rewrite
	FullRegexMatch = "fullregex"
	// NoOpMatch placeholder that does nothing
	NoOpMatch = "noop"
	namePart  = "name"
	dataPart  = "data"
	bothPart  = "both"
)

const (
	// RewriteIgnored is returned when rewrite is not done on request.
	RewriteIgnored Result = iota
	// RewriteDone is returned when rewrite is done on request.
	RewriteDone
)

// Result is the result of a rewrite
type Result int

// Rule describes a rewrite rule.
type Rule interface {
	Rewrite(ctx context.Context, state request.Request) Result
	Sub(n string) string
	RRPart() string
}

type exactRule struct {
	rrPart string
	From   string
	To     string
}

type prefixRule struct {
	rrPart      string
	Prefix      string
	Replacement string
}

type suffixRule struct {
	rrPart      string
	Suffix      string
	Replacement string
}

type substringRule struct {
	rrPart      string
	Substring   string
	Replacement string
}

type regexRule struct {
	rrPart      string
	Pattern     *regexp.Regexp
	Replacement string
}

type fullRegexRule struct {
	rrPart      string
	Pattern     *regexp.Regexp
	Replacement string
}

type noOpRule struct {
	rrPart string
	From   string
}

type nameRule struct {
	RuleType   string
	Rule       Rule
	RRType     uint16
	RRClass    uint16
	Answers    []Rule
	Additional []Rule
	Authority  []Rule
}

func (rule *exactRule) Rewrite(ctx context.Context, state request.Request) Result {
	if rule.From == state.Name() {
		state.Req.Question[0].Name = rule.To
		state.Req.RecursionDesired = true
		return RewriteDone
	}
	return RewriteIgnored
}

func (rule *exactRule) Sub(n string) (o string) {
	if rule.From == n {
		o = rule.To
	}
	return
}

func (rule *exactRule) RRPart() string {
	return rule.rrPart
}

func (rule *prefixRule) Rewrite(ctx context.Context, state request.Request) Result {
	if strings.HasPrefix(state.Name(), rule.Prefix) {
		state.Req.Question[0].Name = rule.Replacement + strings.TrimPrefix(state.Name(), rule.Prefix)
		state.Req.RecursionDesired = true
		return RewriteDone
	}
	return RewriteIgnored
}

func (rule *prefixRule) Sub(n string) (o string) {
	if strings.HasPrefix(n, rule.Prefix) {
		o = rule.Replacement + strings.TrimPrefix(n, rule.Prefix)
	}
	return
}

func (rule *prefixRule) RRPart() string {
	return rule.rrPart
}

func (rule *suffixRule) Rewrite(ctx context.Context, state request.Request) Result {
	if strings.HasSuffix(state.Name(), rule.Suffix) {
		state.Req.Question[0].Name = strings.TrimSuffix(state.Name(), rule.Suffix) + rule.Replacement
		state.Req.RecursionDesired = true
		return RewriteDone
	}
	return RewriteIgnored
}

func (rule *suffixRule) Sub(n string) (o string) {
	if strings.HasSuffix(n, rule.Suffix) {
		o = strings.TrimSuffix(n, rule.Suffix) + rule.Replacement
	}
	return
}

func (rule *suffixRule) RRPart() string {
	return rule.rrPart
}

func (rule *substringRule) Rewrite(ctx context.Context, state request.Request) Result {
	if strings.Contains(state.Name(), rule.Substring) {
		state.Req.Question[0].Name = strings.Replace(state.Name(), rule.Substring, rule.Replacement, -1)
		state.Req.RecursionDesired = true
		return RewriteDone
	}
	return RewriteIgnored
}

func (rule *substringRule) Sub(n string) (o string) {
	if strings.Contains(n, rule.Substring) {
		o = strings.Replace(n, rule.Substring, rule.Replacement, -1)
	}
	return
}

func (rule *substringRule) RRPart() string {
	return rule.rrPart
}

func (rule *regexRule) Rewrite(ctx context.Context, state request.Request) Result {
	regexGroups := rule.Pattern.FindStringSubmatch(state.Name())
	if len(regexGroups) == 0 {
		return RewriteIgnored
	}
	s := rule.Replacement
	for groupIndex, groupValue := range regexGroups {
		groupIndexStr := "{" + strconv.Itoa(groupIndex) + "}"
		if strings.Contains(s, groupIndexStr) {
			s = strings.Replace(s, groupIndexStr, groupValue, -1)
		}
	}
	state.Req.Question[0].Name = s
	state.Req.RecursionDesired = true
	return RewriteDone
}

func (rule *regexRule) Sub(n string) (o string) {
	regexGroups := rule.Pattern.FindStringSubmatch(n)
	if len(regexGroups) == 0 {
		return
	}
	s := rule.Replacement
	for groupIndex, groupValue := range regexGroups {
		groupIndexStr := "{" + strconv.Itoa(groupIndex) + "}"
		if strings.Contains(s, groupIndexStr) {
			s = strings.Replace(s, groupIndexStr, groupValue, -1)
			o = s
			break
		}
	}
	return
}

func (rule *regexRule) RRPart() string {
	return rule.rrPart
}

func (rule *fullRegexRule) Rewrite(ctx context.Context, state request.Request) Result {
	if rule.Pattern.MatchString(state.Name()) {
		state.Req.Question[0].Name = plugin.Name(rule.Replacement).Normalize()
		state.Req.RecursionDesired = true
		return RewriteDone
	}
	return RewriteIgnored
}

func (rule *fullRegexRule) Sub(n string) (o string) {
	if rule.Pattern.MatchString(n) {
		o = rule.Replacement
	}
	return
}

func (rule *fullRegexRule) RRPart() string {
	return rule.rrPart
}

func (rule *noOpRule) Rewrite(ctx context.Context, state request.Request) Result {
	if strings.HasSuffix(state.Name(), rule.From) {
		state.Req.RecursionDesired = true
		return RewriteDone
	}
	return RewriteIgnored
}

func (rule *noOpRule) Sub(n string) (o string) {
	if strings.HasSuffix(n, rule.From) {
		o = rule.From
	}
	return
}

func (rule *noOpRule) RRPart() string {
	return rule.rrPart
}

func (rule *nameRule) Rewrite(ctx context.Context, state request.Request) Result {
	if rule.RRClass != dns.ClassANY && state.QClass() != rule.RRClass {
		return RewriteIgnored
	}
	if rule.RRType != dns.TypeANY && state.QType() != rule.RRType {
		return RewriteIgnored
	}
	return rule.Rule.Rewrite(ctx, state)
}

func (rule *nameRule) Sub(n, t, pt string) (o string) {
	var rules []Rule
	switch t {
	case answerRule:
		rules = rule.Answers
	case authorityRule:
		rules = rule.Authority
	case additionalRule:
		rules = rule.Additional
	default:
		return
	}
	for _, rule := range rules {
		if rule.RRPart() != bothPart && pt != rule.RRPart() {
			continue
		}
		if s := rule.Sub(n); s != "" {
			o = s
			break
		}
	}
	return
}

func newRule(args ...string) (Rule, error) {
	var numArgs int
	var rrPart, mt, from, to string

	numArgs = len(args)

	if numArgs < 3 {
		return nil, fmt.Errorf("Atleast 3 arguments required")
	}

	if numArgs > 3 {
		rrPart = args[0]
		mt = args[1]
		from = args[2]
		to = args[3]
	} else {
		rrPart = namePart
		mt = args[0]
		from = args[1]
		to = args[2]
	}

	if rrPart != bothPart && rrPart != namePart && rrPart != dataPart {
		return nil, fmt.Errorf("Only (name, data, both) RR parts are supported, received: %s", rrPart)
	}

	if mt == ExactMatch || mt == SuffixMatch {
		if !hasClosingDot(from) {
			from = from + "."
		}
		if !hasClosingDot(to) {
			to = to + "."
		}
	}
	switch mt {
	case ExactMatch:
		return &exactRule{
			rrPart,
			from,
			to,
		}, nil
	case PrefixMatch:
		return &prefixRule{
			rrPart,
			from,
			to,
		}, nil
	case SuffixMatch:
		return &suffixRule{
			rrPart,
			from,
			to,
		}, nil
	case SubstringMatch:
		return &substringRule{
			rrPart,
			from,
			to,
		}, nil
	case RegexMatch:
		rewriteQuestionFromPattern, err := isValidRegexPattern(from, to)
		if err != nil {
			return nil, err
		}
		rewriteQuestionTo := plugin.Name(to).Normalize()
		return &regexRule{
			rrPart,
			rewriteQuestionFromPattern,
			rewriteQuestionTo,
		}, nil
	case FullRegexMatch:
		rewriteQuestionFromPattern, err := regexp.Compile(from)
		if err != nil {
			return nil, err
		}
		return &fullRegexRule{
			rrPart,
			rewriteQuestionFromPattern,
			to,
		}, nil
	case NoOpMatch:
		return &noOpRule{
			rrPart,
			from,
		}, nil
	default:
		return nil, fmt.Errorf("Only exact, prefix, suffix, substring, regex, fullregex, noop rule types are supported, received: %s", mt)
	}
}

func hasClosingDot(s string) bool {
	if strings.HasSuffix(s, ".") {
		return true
	}
	return false
}

func getSubExprUsage(s string) int {
	subExprUsage := 0
	for i := 0; i <= 100; i++ {
		if strings.Contains(s, "{"+strconv.Itoa(i)+"}") {
			subExprUsage++
		}
	}
	return subExprUsage
}

func isValidRegexPattern(rewriteFrom, rewriteTo string) (*regexp.Regexp, error) {
	rewriteFromPattern, err := regexp.Compile(rewriteFrom)
	if err != nil {
		return nil, fmt.Errorf("invalid regex matching pattern: %s", rewriteFrom)
	}
	if getSubExprUsage(rewriteTo) > rewriteFromPattern.NumSubexp() {
		return nil, fmt.Errorf("the rewrite regex pattern (%s) uses more subexpressions than its corresponding matching regex pattern (%s)", rewriteTo, rewriteFrom)
	}
	return rewriteFromPattern, nil
}

func validName(s string) bool {
	_, ok := dns.IsDomainName(s)
	if !ok {
		return false
	}
	if len(dns.Name(s).String()) > 255 {
		return false
	}

	return true
}
