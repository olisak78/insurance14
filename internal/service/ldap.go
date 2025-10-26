package service

import (
	"crypto/tls"
	"strings"
	"time"

	"developer-portal-backend/internal/config"

	"github.com/go-ldap/ldap/v3"
)

// LDAPUser represents a subset of LDAP user attributes returned by the search
type LDAPUser struct {
	DN          string `json:"dn"`
	DisplayName string `json:"displayName"`
	Mobile      string `json:"mobile"`
	SN          string `json:"sn"`
	Name        string `json:"name"`
	Mail        string `json:"mail"`
	GivenName   string `json:"givenName"`
}

// LDAPService provides methods to interact with LDAP
type LDAPService struct {
	cfg *config.Config
}

// NewLDAPService creates a new LDAP service
func NewLDAPService(cfg *config.Config) *LDAPService {
	return &LDAPService{cfg: cfg}
}

// SearchUsersByCN searches users by common name (cn prefix match)
func (s *LDAPService) SearchUsersByCN(cn string) ([]LDAPUser, error) {
	addr := s.cfg.LDAPHost + ":" + s.cfg.LDAPPort

	// Establish TLS connection to LDAP server
	l, err := ldap.DialTLS("tcp", addr, &tls.Config{InsecureSkipVerify: s.cfg.LDAPInsecureSkipVerify})
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// Set timeout
	if s.cfg.LDAPTimeoutSec > 0 {
		l.SetTimeout(time.Duration(s.cfg.LDAPTimeoutSec) * time.Second)
	}

	// Bind with configured credentials
	if err := l.Bind(s.cfg.LDAPBindDN, s.cfg.LDAPBindPW); err != nil {
		return nil, err
	}

	// Build search request
	filter := "(cn=" + ldap.EscapeFilter(cn) + "*)"
	attrs := []string{"displayName", "mobile", "sn", "name", "mail", "givenName"}
	// Adjust base DN prefix based on first letter of search string (case-insensitive)
	baseDN := s.cfg.LDAPBaseDN
	if len(cn) > 0 {
		switch strings.ToLower(string(cn[0])) {
		case "i":
			baseDN = "OU=I," + baseDN
		case "d":
			baseDN = "OU=D," + baseDN
		case "c":
			baseDN = "OU=C," + baseDN
		}
	}

	req := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		s.cfg.LDAPTimeoutSec,
		false,
		filter,
		attrs,
		nil,
	)

	// Execute search
	res, err := l.Search(req)
	if err != nil {
		return nil, err
	}

	// Map results
	out := make([]LDAPUser, 0, len(res.Entries))
	for _, e := range res.Entries {
		get := func(a string) string { return e.GetAttributeValue(a) }
		out = append(out, LDAPUser{
			DN:          e.DN,
			DisplayName: get("displayName"),
			Mobile:      get("mobile"),
			SN:          get("sn"),
			Name:        get("name"),
			Mail:        get("mail"),
			GivenName:   get("givenName"),
		})
	}

	return out, nil
}
