package main

import (
  "encoding/json"
  "fmt"
  "os"
  "net/http"
  "time"
  "io/ioutil"
  "bytes"
  "bufio"
  "strings"
  "flag"
  "runtime"
  // for getting a password without echoing it
  "golang.org/x/crypto/ssh/terminal"
  "syscall"
)

func main() {
  urlAuth := "https://pouta.csc.fi:5001/v3/auth/tokens"

  projPtr := flag.String("p", "", "project name")
  userPtr := flag.String("u", "", "username")
  helpPtr := flag.Bool("h", false, "help")
  debugPtr := flag.Bool("d", false, "debug")
  flag.Parse()
  project := *projPtr
  user := *userPtr

  if *helpPtr {
    fmt.Println("Usage: " + os.Args[0] + " [-h] [-d] [-u=username] [-p=projectname]")
    fmt.Println("       -h = this help")
    fmt.Println("       -d = debug, shows names to be sent and also the response header")
    os.Exit(0)
  }

  // for missing values ask the user

  if project == "" {
    project = askStr("project name")
    if project == "" { os.Exit(1) }
  }
  if *debugPtr { fmt.Println("DEBUG project=" + project) }

  if user == "" {
    user = os.Getenv("LOGNAME")
  }
  if user == "" {
    user = askStr("username")
    if user == "" { os.Exit(1) }
  }
  if *debugPtr { fmt.Println("DEBUG user=" + user) }

  fmt.Print("Enter password for user \"", user, "\": ")
  bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
  fmt.Println("")
  if err != nil {
    fmt.Println("Error reading user input")
    os.Exit(1)
  }
  passwd := string(bytePassword)
  if runtime.GOOS == "windows" {
    passwd = strings.TrimSuffix(passwd, "\r\n")
  } else {
    passwd = strings.TrimSuffix(passwd, "\n")
  }

  // create authentication json

  jsonStr := "{ \"auth\": { \"identity\": { \"methods\": [ \"password\" ], \"password\": { \"user\": { \"id\": \"" +
    user +
    "\", \"password\": \"" +
    passwd +
    "\" } } }, \"scope\": { \"project\": { \"domain\": { \"id\": \"default\" }, \"name\": \"" +
    project +
    "\" } } } }"

  // authenticate

  req, err := http.NewRequest("POST", urlAuth, bytes.NewBuffer([]byte(jsonStr)))
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  client.Timeout = time.Second * 15
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()

  if resp.Status != "201 Created" {
    fmt.Println("Authentication failed. HTTP response code:", resp.Status)
    os.Exit(1)
  }

  // parse the authentication token from response headers

  if *debugPtr { fmt.Print("DEBUG resp.Header="); fmt.Println(resp.Header) }
  allasToken := ""
  if isset(resp.Header["X-Subject-Token"], 0) {
    allasToken = resp.Header["X-Subject-Token"][0]
  } else {
    fmt.Println("Error in the server response. Authentication token not found")
    os.Exit(1)
  }

  // parse the storage url from response body json

  body, _ := ioutil.ReadAll(resp.Body)
  var result map[string]interface{}
  json.Unmarshal((body), &result)
  allasEndpoint := parseResult(result)

  if allasToken != "" && allasEndpoint != "" {
    if runtime.GOOS == "windows" {
      fmt.Println("Run these commands in your command propmpt:")
      fmt.Println("set OS_AUTH_TOKEN=" + allasToken)
      fmt.Println("set OS_STORAGE_URL=" + allasEndpoint)
    } else {
      fmt.Println("Run these commands in your shell:")
      fmt.Println("export OS_AUTH_TOKEN=" + allasToken)
      fmt.Println("export OS_STORAGE_URL=" + allasEndpoint)
    }
  }
}

func isset(array []string, i int) bool {
    return (len(array) > i)
}

func askStr(name string) string {
  value := ""
  reader := bufio.NewReader(os.Stdin)
  fmt.Print("Enter " + name + ": ")
  value, _ = reader.ReadString('\n')
  if runtime.GOOS == "windows" {
    value = strings.TrimSuffix(value, "\r\n")
  } else {
    value = strings.TrimSuffix(value, "\n")
  }
  return value
}

// there may be easier way to parse the result, but as long as this works...

func parseResult(result map[string]interface{}) string {
  token := result["token"].(map[string]interface{})
  catalog := token["catalog"].([]interface{})
  url := ""

  for key, val := range catalog {
    key = key
    for key2, val2 := range val.(map[string]interface{}) {
      switch val2.(type) {
        case interface{}:
          if key2 == "name" && val2 == "swift" {
            url = parseEntry(val.(map[string]interface{}))
            if url != "" {
              return url
            }
          }
      }
    }
  }
  return url
}

func parseEntry(entry map[string]interface{}) string {
  url := ""
  for key, val := range entry {
    key = key
    switch val.(type) {
      case []interface{}:
        for key2, val2 := range val.([]interface{}) {
          key2 = key2
          url = parseEndpoint(val2.(map[string]interface{}))
          if url != "" {
            return url
          }
        }
    }
  }
  return ""
}

func parseEndpoint(endpoint map[string]interface{}) string {
  url := ""
  if endpoint["interface"] == "public" {
    url = endpoint["url"].(string)
  }
  return url
}

