module github.com/adibhanna/modbus-go

go 1.25

require go.bug.st/serial v1.6.2

require (
	github.com/creack/goselect v0.1.2 // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
)

// Retract stale versions cached by the Go module proxy from deleted/re-tagged git tags.
// These versions point to early commits missing features like SetAutoReconnect,
// TLS, RTU-over-TCP, UDP transports, and high-level data types.
// Use v1.3.0 or later instead.
retract (
	v1.2.0 // Points to early commit, missing features added after v1.1.0 tag
	v1.1.0 // Points to early commit, missing high-level data types and transports
	v1.0.2 // Points to early commit, same as proxy v1.0.0
	v1.0.0 // Points to early commit, missing JSON configuration system
)
