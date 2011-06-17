package flickgo

import (
  "crypto/md5"
  "fmt"
  "hash"
  "testing"
)

const (
  secret = "sf97838dijd"
)

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
  write(m, "xyz" + "xyz")
  expected := fmt.Sprintf("https://secure.flickr.com/services?" +
      "123=98765&abc=abc+def&xyz=xyz&api_sig=%x", m.Sum())

  actual, err := Sign(secret, args)
  if err != nil {
    t.Errorf("unexpcted error: %v", err)
  }
  if expected != actual.String() {
    t.Errorf("expcted: %q, found %q", expected, actual)
  }
}
