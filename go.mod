module github.com/synprosolutions/haproxy-healthcheck

go 1.11

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/coreos/go-systemd v0.0.0-00010101000000-000000000000
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/shirou/gopsutil v2.20.4+incompatible
	github.com/stretchr/testify v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20200523222454-059865788121 // indirect
)

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
