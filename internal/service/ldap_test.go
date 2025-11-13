package service

import (
	"crypto/tls"
	"errors"
	"testing"
	"time"

	"developer-portal-backend/internal/config"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
)

// fakeLDAPClient implements ldapClient for testing
type fakeLDAPClient struct {
	bindErr           error
	searchErr         error
	searchRes         *ldap.SearchResult
	receivedSearchReq *ldap.SearchRequest

	setTimeoutCalled bool
	timeoutValue     time.Duration

	closed bool
}

func (f *fakeLDAPClient) Bind(username, password string) error {
	return f.bindErr
}

func (f *fakeLDAPClient) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	f.receivedSearchReq = searchRequest
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	if f.searchRes != nil {
		return f.searchRes, nil
	}
	return &ldap.SearchResult{Entries: []*ldap.Entry{}}, nil
}

func (f *fakeLDAPClient) Close() error {
	f.closed = true
	return nil
}

func (f *fakeLDAPClient) SetTimeout(d time.Duration) {
	f.setTimeoutCalled = true
	f.timeoutValue = d
}

func makeConfig() *config.Config {
	return &config.Config{
		LDAPHost:               "ldap.example.com",
		LDAPPort:               "636",
		LDAPBindDN:             "CN=John Doe,OU=Users,DC=example,DC=com",
		LDAPBindPW:             "SuperSecret123",
		LDAPBaseDN:             "DC=example,DC=com",
		LDAPInsecureSkipVerify: true,
		LDAPTimeoutSec:         5,
	}
}

func TestLDAP_SearchUsersByCN_DialError(t *testing.T) {
	orig := dialLDAP
	defer func() { dialLDAP = orig }()

	dialLDAP = func(network, addr string, cfg *tls.Config) (ldapClient, error) {
		return nil, errors.New("dial failed")
	}

	svc := NewLDAPService(makeConfig())
	res, err := svc.SearchUsersByCN("john")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "dial failed")
}

func TestLDAP_SearchUsersByCN_BindError(t *testing.T) {
	orig := dialLDAP
	defer func() { dialLDAP = orig }()

	fc := &fakeLDAPClient{bindErr: errors.New("bind failed")}
	dialLDAP = func(network, addr string, cfg *tls.Config) (ldapClient, error) {
		return fc, nil
	}

	cfg := makeConfig()
	svc := NewLDAPService(cfg)
	res, err := svc.SearchUsersByCN("john")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "bind failed")
	assert.True(t, fc.closed, "client should be closed via defer")
	assert.True(t, fc.setTimeoutCalled, "SetTimeout should be called")
	assert.Equal(t, time.Duration(cfg.LDAPTimeoutSec)*time.Second, fc.timeoutValue)
}

func TestLDAP_SearchUsersByCN_SearchError(t *testing.T) {
	orig := dialLDAP
	defer func() { dialLDAP = orig }()

	fc := &fakeLDAPClient{
		searchErr: errors.New("search failed"),
	}
	dialLDAP = func(network, addr string, cfg *tls.Config) (ldapClient, error) {
		return fc, nil
	}

	svc := NewLDAPService(makeConfig())
	res, err := svc.SearchUsersByCN("john")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "search failed")
	assert.NotNil(t, fc.receivedSearchReq, "Search should receive a request")
}

func TestLDAP_SearchUsersByCN_Success_Mapping_And_BaseDN_Prefix(t *testing.T) {
	type tc struct {
		name       string
		cn         string
		expectOU   string // "", "OU=I,", "OU=D,", "OU=C,"
		entry      *ldap.Entry
	}
	baseDN := "DC=example,DC=com"

	makeEntry := func() *ldap.Entry {
		return &ldap.Entry{
			DN: "CN=John Doe,OU=Users,DC=example,DC=com",
			Attributes: []*ldap.EntryAttribute{
				{Name: "displayName", Values: []string{"John Doe"}},
				{Name: "mobile", Values: []string{"+1-555-1234"}},
				{Name: "sn", Values: []string{"Doe"}},
				{Name: "name", Values: []string{"jdoe"}},
				{Name: "mail", Values: []string{"jdoe@example.com"}},
				{Name: "givenName", Values: []string{"John"}},
			},
		}
	}

	tests := []tc{
		{name: "Default base DN", cn: "bob", expectOU: "", entry: makeEntry()},
		{name: "OU=I prefix", cn: "igor", expectOU: "OU=I,", entry: makeEntry()},
		{name: "OU=D prefix", cn: "dan", expectOU: "OU=D,", entry: makeEntry()},
		{name: "OU=C prefix", cn: "chris", expectOU: "OU=C,", entry: makeEntry()},
}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := dialLDAP
			defer func() { dialLDAP = orig }()

			fc := &fakeLDAPClient{
				searchRes: &ldap.SearchResult{
					Entries: []*ldap.Entry{tt.entry},
				},
			}
			dialLDAP = func(network, addr string, cfg *tls.Config) (ldapClient, error) {
				return fc, nil
			}

			cfg := makeConfig()
			cfg.LDAPBaseDN = baseDN
			svc := NewLDAPService(cfg)

			out, err := svc.SearchUsersByCN(tt.cn)
			assert.NoError(t, err)
			assert.Len(t, out, 1)
			user := out[0]
			assert.Equal(t, tt.entry.DN, user.DN)
			assert.Equal(t, "John Doe", user.DisplayName)
			assert.Equal(t, "+1-555-1234", user.Mobile)
			assert.Equal(t, "Doe", user.SN)
			assert.Equal(t, "jdoe", user.Name)
			assert.Equal(t, "jdoe@example.com", user.Mail)
			assert.Equal(t, "John", user.GivenName)

			if assert.NotNil(t, fc.receivedSearchReq) {
				expectedBase := tt.expectOU + baseDN
				assert.Equal(t, expectedBase, fc.receivedSearchReq.BaseDN)
				// also validate filter built correctly with wildcard suffix and escaped value
				assert.Equal(t, "(cn="+ldap.EscapeFilter(tt.cn)+"*)", fc.receivedSearchReq.Filter)
			}
		})
	}
}

func TestLDAP_SearchUsersByCN_TimeoutZero_DoesNotSet(t *testing.T) {
	orig := dialLDAP
	defer func() { dialLDAP = orig }()

	fc := &fakeLDAPClient{}
	dialLDAP = func(network, addr string, cfg *tls.Config) (ldapClient, error) {
		return fc, nil
	}

	cfg := makeConfig()
	cfg.LDAPTimeoutSec = 0 // zero timeout should not call SetTimeout
	svc := NewLDAPService(cfg)
	_, _ = svc.SearchUsersByCN("alice")

	assert.False(t, fc.setTimeoutCalled, "SetTimeout should not be called when timeout is 0")
}
