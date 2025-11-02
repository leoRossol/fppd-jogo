package main

import (
  "fmt"
  "net/rpc"
  "os"
  "strconv"
)

func main() {
  addr := "127.0.0.1:12345"
  if a := os.Getenv("RPC_ADDR"); a != "" { addr = a }
  client, err := rpc.Dial("tcp", addr)
  if err != nil {
    fmt.Println("dial error:", err)
    return
  }
  seqEnv := os.Getenv("SEQ")
  seq := int64(1)
  if seqEnv != "" {
    s, _ := strconv.ParseInt(seqEnv, 10, 64)
    seq = s
  }
  clientID := os.Getenv("CLIENTID")
  if clientID == "" {
    fmt.Println("CLIENTID env required")
    return
  }
  args := CommandArgs{ClientID: clientID, Seq: seq, Cmd: "UPDATE_POS", Payload: map[string]interface{}{"x": 5, "y": 5, "lives": 3}}
  var reply CommandReply
  if err := client.Call("GameServer.SendCommand", &args, &reply); err != nil {
    fmt.Println("call error:", err)
  } else {
    fmt.Printf("reply: %+v\n", reply)
  }
}
