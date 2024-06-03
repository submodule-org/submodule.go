package mhttp_test

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/submodule-org/submodule.go"
	"github.com/submodule-org/submodule.go/meta/mhttp"
)

type Hello struct{}

func (h *Hello) AdaptToHTTPHandler(m *http.ServeMux) {
	m.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
}

var HelloRoute = submodule.Resolve(&Hello{})

type MHTTPSuite struct {
	suite.Suite
	*require.Assertions
	scope submodule.Scope

	server *http.Server
	e      error
	port   int
}

func (s *MHTTPSuite) SetupSubTest() {
	s.scope = submodule.CreateScope()
	_, e := HelloRoute.SafeResolveWith(s.scope)
	s.Require().Nil(e)
}

func (s *MHTTPSuite) TearDownSubTest() {
	s.server, s.e = mhttp.Server.SafeResolveWith(s.scope)
	s.Require().Nil(s.e)

	go func() {
		s.server.ListenAndServe()
	}()

	defer func() {
		s.server.Close()
	}()
	defer s.scope.Dispose()
	defer mhttp.Reset()

	time.Sleep(200 * time.Millisecond)
	r, e := http.Get(fmt.Sprintf("http://localhost:%d/hello", s.port))

	s.Require().Nil(e)
	s.Require().Equal(200, r.StatusCode)

	body, e := io.ReadAll(r.Body)
	s.Require().Nil(e)
	s.Require().Equal("hello", string(body))

}

func (s *MHTTPSuite) TestCanChangePort() {
	s.port = 28001

	mhttp.AlterConfig(func(c *mhttp.ServerConfig) {
		c.Addr = fmt.Sprintf(":%d", s.port)
	})
}

func (s *MHTTPSuite) TestCanStartServer() {
}

func TestMHTTP(t *testing.T) {
	suite.Run(t, &MHTTPSuite{
		port: 8080,
	})
}
