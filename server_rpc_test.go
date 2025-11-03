package main

import (
    "net"
    "net/rpc"
    "testing"
    "time"
)

// TestExactlyOnceViaRPC inicia um servidor RPC em porta dinâmica e valida que
// reenviar o mesmo ClientID+Seq retorna a mesma resposta (deduplicação)
func TestExactlyOnceViaRPC(t *testing.T) {
    gs := NewGameServer()

    // registrar e iniciar listener em porta aleatória
    if err := rpc.Register(gs); err != nil {
        t.Fatalf("rpc.Register error: %v", err)
    }
    l, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        t.Fatalf("listen error: %v", err)
    }
    defer l.Close()

    go rpc.Accept(l)

    // cria cliente RPC
    client, err := rpc.Dial("tcp", l.Addr().String())
    if err != nil {
        t.Fatalf("dial error: %v", err)
    }
    defer client.Close()

    // envia REGISTER
    args := CommandArgs{ClientID: "test-client", Seq: 1, Cmd: "REGISTER", Payload: RegisterPayload{Name: "tester"}}
    var reply CommandReply
    if err := client.Call("GameServer.SendCommand", &args, &reply); err != nil {
        t.Fatalf("SendCommand error: %v", err)
    }
    if !reply.Applied {
        t.Fatalf("expected applied=true, got reply=%+v", reply)
    }

    // reenvia mesmo comando (mesmo seq) -> deve retornar o mesmo reply
    var reply2 CommandReply
    if err := client.Call("GameServer.SendCommand", &args, &reply2); err != nil {
        t.Fatalf("SendCommand second call error: %v", err)
    }
    if reply2 != reply {
        t.Fatalf("expected identical replies on duplicate: got %+v and %+v", reply, reply2)
    }

    // verificar GetState retorna 1 jogador
    var st StateReply
    if err := client.Call("GameServer.GetState", &ClientIDArgs{ClientID: "test-client", Now: time.Now()}, &st); err != nil {
        t.Fatalf("GetState error: %v", err)
    }
    if len(st.Players) != 1 {
        t.Fatalf("expected 1 player in state, got %d", len(st.Players))
    }
}
