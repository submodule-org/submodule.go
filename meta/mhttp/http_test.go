package mhttp_test

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"

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
	e := mhttp.ResolveRoutesIn(s.scope, HelloRoute)
	s.Require().Nil(e)
}

func (s *MHTTPSuite) TearDownSubTest() {
	s.server, s.e = mhttp.Server.SafeResolveWith(s.scope)
	s.Require().Nil(s.e)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		s.server.ListenAndServe()
	}()

	defer func() {
		s.server.Close()
	}()
	defer s.scope.Dispose()
	defer mhttp.Reset()

	wg.Wait()
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
