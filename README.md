# redactif [![GoDoc](https://godoc.org/github.com/bradleyjkemp/redactif?status.svg)](https://godoc.org/github.com/bradleyjkemp/redactif)
Go library for zeroing marked fields in an arbitrary data structure.

This was originally written to help when Snapshot testing structures that contain unpredictable fields e.g.
```go
type result struct {
    // unpredictable field that is tricky to snapshot
    timestamp time.Time `redactif:"snapshot"`

    // normal field
    value string
}

// In your snapshot test use redactif.Redact()
// to easily zero the tricky fields
r := someFunctionReturnsResult()
redactif.Redact(r, "snapshot")
cupaloy.SnapshotT(t, r)
```

However you could also use this to implement things like preventing passwords from appearing in logs e.g.
```go
type user struct {
    username string
    email string `redactif:"!internal"`
    passwordHash []byte `redactif:"log"`
}

// When logging:
u := getUserDetails()
log.Debug("User logged in", redactif.Redact(u, "log", "internal"))
// And the sensitive passwordHash field will have been removed
```
