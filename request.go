package flickgo

import (
  "crypto/md5"
  "fmt"
  "http"
  "os"
  "sort"
)

const (
  service = "https://secure.flickr.com/services"
)

func keys(m map[string]string) sort.StringArray {
  keys := make([]string, len(m))
  i := 0
  for k, _ := range m {
    keys[i] = k
    i++
  }
  return keys
}

func Sign(secret string, args map[string]string) (url *http.URL, err os.Error) {
  ks := keys(args)
  ks.Sort()
  s := service + "?"
  m := md5.New()
  m.Write([]byte(secret))
  for _, k := range ks {
    value := http.URLEscape(args[k])
    s += fmt.Sprintf("%s=%s&", k, value)
    m.Write([]byte(k + value))
  }
  s += fmt.Sprintf("api_sig=%x", m.Sum())
  return http.ParseURL(s)
}
