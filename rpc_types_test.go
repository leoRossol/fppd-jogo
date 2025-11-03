package main

import (
    "bytes"
    "encoding/gob"
    "reflect"
    "testing"
)

func TestGobEncodeDecode_UpdatePosPayload(t *testing.T) {
    orig := CommandArgs{
        ClientID: "client-123",
        Seq:      42,
        Cmd:      "UPDATE_POS",
        Payload:  UpdatePosPayload{X: 10, Y: 20, Lives: 3},
    }

    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    if err := enc.Encode(orig); err != nil {
        t.Fatalf("encode failed: %v", err)
    }

    var decoded CommandArgs
    dec := gob.NewDecoder(&buf)
    if err := dec.Decode(&decoded); err != nil {
        t.Fatalf("decode failed: %v", err)
    }

    if decoded.ClientID != orig.ClientID || decoded.Seq != orig.Seq || decoded.Cmd != orig.Cmd {
        t.Fatalf("metadata mismatch: got=%v want=%v", decoded, orig)
    }

    // Verificar tipo e valor do payload
    upd, ok := decoded.Payload.(UpdatePosPayload)
    if !ok {
        t.Fatalf("decoded payload has wrong type: %T", decoded.Payload)
    }
    if !reflect.DeepEqual(upd, orig.Payload) {
        t.Fatalf("payload mismatch: got=%v want=%v", upd, orig.Payload)
    }
}

func TestGobEncodeDecode_RegisterPayload(t *testing.T) {
    orig := CommandArgs{
        ClientID: "client-xyz",
        Seq:      1,
        Cmd:      "REGISTER",
        Payload:  RegisterPayload{Name: "Alice", X: 1, Y: 2},
    }

    var buf bytes.Buffer
    if err := gob.NewEncoder(&buf).Encode(orig); err != nil {
        t.Fatalf("encode failed: %v", err)
    }

    var decoded CommandArgs
    if err := gob.NewDecoder(&buf).Decode(&decoded); err != nil {
        t.Fatalf("decode failed: %v", err)
    }

    reg, ok := decoded.Payload.(RegisterPayload)
    if !ok {
        t.Fatalf("decoded payload has wrong type: %T", decoded.Payload)
    }
    if !reflect.DeepEqual(reg, orig.Payload) {
        t.Fatalf("payload mismatch: got=%v want=%v", reg, orig.Payload)
    }
}

func TestValidateUpdatePos(t *testing.T) {
    valid := UpdatePosPayload{X: 0, Y: 0, Lives: 1}
    if err := ValidateUpdatePos(valid); err != nil {
        t.Fatalf("valid payload considered invalid: %v", err)
    }

    invalid := UpdatePosPayload{X: -1, Y: 0, Lives: 1}
    if err := ValidateUpdatePos(invalid); err == nil {
        t.Fatalf("invalid payload (negative X) considered valid")
    }
}

func TestValidateRegister(t *testing.T) {
    valid := RegisterPayload{Name: "Bob", X: 0, Y: 0}
    if err := ValidateRegister(valid); err != nil {
        t.Fatalf("valid register considered invalid: %v", err)
    }

    invalid := RegisterPayload{Name: "", X: 0, Y: 0}
    if err := ValidateRegister(invalid); err == nil {
        t.Fatalf("invalid register (empty name) considered valid")
    }
}
