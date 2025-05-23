module github.com/d-protocol/pokerlib

go 1.19

require (
	github.com/d-protocol/pokertable v0.0.0-20250522235840-06d931cd5fc5
	github.com/d-protocol/syncsaga v0.0.0-20250522234550-88c17af8cd6c
	github.com/d-protocol/timebank v0.0.0-20250522234202-39608dfc16ba
	github.com/google/uuid v1.3.0
	github.com/nats-io/nats-server/v2 v2.9.20
	github.com/nats-io/nats.go v1.28.0
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.16.5 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/nats-io/jwt/v2 v2.4.1 // indirect
	github.com/nats-io/nkeys v0.4.4 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/thoas/go-funk v0.9.3 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Use local replacements instead of trying to fetch from remote repositories
replace (
	github.com/d-protocol/pokerface => ../pokerface
	github.com/d-protocol/pokertable => ../pokertable
)
