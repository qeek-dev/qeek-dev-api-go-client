package account

import (
	"testing"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	. "gopkg.in/check.v1"
)

const (
	// remember to change it to a valid token to run test
	AccessToken = ""
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct {
	c *Service
}

func (s *MySuite) SetUpTest(c *C) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: AccessToken})
	tc := oauth2.NewClient(ctx, ts)
	s.c = New(tc)
}

var _ = Suite(&MySuite{})

func (s *MySuite) Test_Myqnapcloud_Account_Me(chk *C) {
	res, err := s.c.Me.Get().Do()
	if err != nil {
		chk.Error(err)
	} else {
		chk.Log(res)
	}
}
