// Copyright 2011 Muthukannan T <manki@manki.in>. All Rights Reserved.

package flickgo

import (
  "bytes"
  "crypto/md5"
  "fmt"
  "hash"
  "http"
  "io"
  "os"
  "strings"
  "testing"
)

const (
  apiKey = "87337fd784"
  secret = "sf97838dijd"
)


func assert(t *testing.T, id string, cond bool) {
  if !cond {
    t.Errorf("[%s] assertion failed", id)
  }
}

func assertOK(t *testing.T, id string, err os.Error) {
  if err != nil {
    t.Errorf("[%s] unexpcted error: <%v>", id, err)
  }
}

func assertEq(t *testing.T, id string, expected interface{}, actual interface{}) {
  if expected != actual {
    t.Errorf("[%s] expcted: <%v>, found <%v>", id, expected, actual)
  }
}


func write(h hash.Hash, s string) {
  h.Write([]byte(s))
}

func TestSign(t *testing.T) {
  args := make(map[string]string)
  args["abc"] = "abc def"
  args["xyz"] = "xyz"
  args["123"] = "98765"

  m := md5.New()
  write(m, secret)
  write(m, "123" + "98765")
  write(m, "abc" + "abc+def")
  write(m, "api_key" + "apap983+key")
  write(m, "xyz" + "xyz")
  expected := fmt.Sprintf("http://www.flickr.com/services/srv/?" +
      "123=98765&abc=abc+def&api_key=apap983+key&xyz=xyz&api_sig=%x", m.Sum())

  actual := sign(secret, "apap983 key", "srv", args)
  assertEq(t, "url", expected, actual.String())
}


type fakeBody struct {
  error os.Error
  data []byte
  read bool
}
func (f fakeBody) Read(buf []byte) (int, os.Error) {
  if (currentBody.read) {
    return 0, os.EOF
  }

  for i, b := range f.data {
    buf[i] = b
  }
  currentBody.read = true
  return len(f.data), f.error
}
func (f fakeBody) Close() os.Error {
  return nil
}

// "Methods" of fakeBody take a fakeBody instance _by value_, which means they
// cannot mutate the instance being operated on.  This global reference will be
// set by tests and mutated by fakeBody's methods.  Big time facepalm!
var currentBody fakeBody

type fakeRoundTripper struct {
  err os.Error
  body fakeBody
  getFn func(r *http.Request) (*http.Response, os.Error)
}
func (f fakeRoundTripper) RoundTrip(r *http.Request) (*http.Response, os.Error) {
  return f.getFn(r)
}

func newHTTPClient(getFn func(*http.Request) (*http.Response, os.Error)) *http.Client {
  rt := fakeRoundTripper{getFn: getFn}
  return &http.Client{Transport: rt}
}

func TestFetchHttpGetFails(t *testing.T) {
  url := "http://some.url/?arg=value"
  err := os.NewError("random error")
  getFn := func(r *http.Request) (*http.Response, os.Error) {
    assertEq(t, "url", url, r.URL.String())
    return nil, err
  }
  c := New(apiKey, secret, newHTTPClient(getFn))

  u, _ := http.ParseURL(url)
  resp, e := c.fetch(u)
  assertEq(t, "resp", 0, len(resp))
  assertEq(t, "err", fmt.Sprintf("Get %s: %s", url, err), e.String())
}

func TestFetchReadFails(t *testing.T) {
  url := "http://some.url/?arg=value"
  err := os.NewError("random error")

  body := fakeBody{error: err}
  currentBody = body
  resp := http.Response{Body: body}
  getFn := func(r *http.Request) (*http.Response, os.Error) {
    assertEq(t, "url", url, r.URL.String())
    return &resp, nil
  }
  c := New(apiKey, secret, newHTTPClient(getFn))
  c.readFn = func(r io.Reader) ([]byte, os.Error) {
    return make([]byte, 0), err
  }

  u, _ := http.ParseURL(url)
  _, e := c.fetch(u)
  assertEq(t, "err", err, e)
}

func TestFetchSuccess(t *testing.T) {
  url := "http://some.url/?arg=value"

  expectedData := bytes.NewBufferString("response from Flickr").Bytes()
  body := fakeBody{data: expectedData}
  currentBody = body
  resp := http.Response{Body: body}
  getFn := func(r *http.Request) (*http.Response, os.Error) {
    assertEq(t, "url", url, r.URL.String())
    return &resp, nil
  }
  c := New(apiKey, secret, newHTTPClient(getFn))

  u, _ := http.ParseURL(url)
  actualData, e := c.fetch(u)
  assertOK(t, "fetch", e)
  assert(t, "data", bytes.Equal(expectedData, actualData))
}


func TestAuthURL(t *testing.T) {
  c := New(apiKey, secret, http.DefaultClient)

  u := c.AuthURL(ReadPerm)
  args, err := http.ParseQuery(u.RawQuery)
  assertOK(t, "parseQuery", err)

  for _, key := range []string{"api_key", "perms", "api_sig"} {
    if (len(args[key]) != 1) {
      t.Errorf("Query argument %s has value %v", key, args[key])
    }
  }
  assertEq(t, "api_key", apiKey, args["api_key"][0])
  assertEq(t, "perms", ReadPerm, args["perms"][0])
}

func TestGetTokenURL(t *testing.T) {
  frob := "837cjnei"
  c := New(apiKey, secret, http.DefaultClient)

  u := getTokenURL(c, frob)
  args, err := http.ParseQuery(u.RawQuery)
  assertOK(t, "parseQuery", err)
  assertEq(t, "method", "flickr.auth.getToken", args["method"][0])
  assertEq(t, "frob", frob, args["frob"][0])
  assertEq(t, "api_key", apiKey, args["api_key"][0])
  assertEq(t, "api_sig", 1, len(args["api_sig"]))
}

func TestGetTokenApiFailure(t *testing.T) {
  jsonStr := `jsonFlickrApi({
    "stat": "fail",
    "code": 97,
    "message": "Missing signature"
  })`
  jsonBytes := bytes.NewBufferString(jsonStr).Bytes()
  body := fakeBody{data: jsonBytes}
  currentBody = body
  resp := http.Response{Body: body}
  getFn := func(r *http.Request) (*http.Response, os.Error) {
    return &resp, nil
  }
  c := New(apiKey, secret, newHTTPClient(getFn))
  _, err := c.GetToken("878243")
  assert(t, "err", err != nil)
  assert(t, "message: " + err.String(),
         strings.Contains(err.String(), "code 97: Missing signature"))
}

func TestGetToken(t *testing.T) {
  jsonStr := `jsonFlickrApi({
    "stat": "ok",
    "auth": {
      "token": {"_content": "121-84669832774"},
      "perms": {"_content": "write"},
      "user": {
        "nsid": "7687633@N01",
        "username": "testuser",
        "fullname": "Test User"
      }
    }
  })`
  jsonBytes := bytes.NewBufferString(jsonStr).Bytes()
  body := fakeBody{data: jsonBytes}
  currentBody = body
  resp := http.Response{Body: body}
  getFn := func(r *http.Request) (*http.Response, os.Error) {
    return &resp, nil
  }
  c := New(apiKey, secret, newHTTPClient(getFn))
  tok, err := c.GetToken("878243")
  assertOK(t, "GetToken", err)
  assertEq(t, "token", "121-84669832774", tok)
}
