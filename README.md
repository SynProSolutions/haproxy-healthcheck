# haproxy-healthcheck

`haproxy-healthcheck` is a tool for usage as [agent-check in haproxy's](https://cbonte.github.io/haproxy-dconv/2.0/configuration.html#5.2-agent-check) server configuration, reporting back CPU idle time (for usage as weight) or the server's administrative state.

The reported integer percentage is similar to what the `vmstat(8)` tool reports in its `id` column when being invoked with a delay of 1 second, like:

    % vmstat 1
    procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----
     r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st
     2  0      0 1939504 213764 4155308    0    0    40    64    5    9  9  5 86  0  0
     0  0      0 1942048 213768 4153104    0    0     0    20 2036 5618  7  7 86  0  0
     0  0      0 1941764 213768 4152592    0    0     0    20 2206 5761  7  8 85  0  0
     [...]

The file `/var/lib/haproxy-healthcheck/statefile` can be used to control the server's administrative state.
If the file exists and contains a supported keyword, its content is reported *instead* of the integer percentage.
The following keywords are currently supported:

* down
* drain
* failed
* maint
* stopped
* ready
* up

**Note:** the optional description strings as well as weights/percentages aren't supported via the file yet.
We currently also don't yet support multiple statements in one single line/file.

**Note:** Do not forget to remove the file to switch back to (live) CPU idle time reporting.

## Usage

*haproxy-healthcheck* relies on [systemd socket activation](https://vincent.bernat.ch/en/blog/2018-systemd-golang-socket-activation).
For testing its integration, you can manually start it to listen on port 2048 via:

    % systemd-socket-activate -l 2048 ./haproxy-healthcheck

## Installation / Build instructions

### Binary packages

Pre-built binary files are available under [releases](https://github.com/SynProSolutions/haproxy-healthcheck/releases).

### Building from source

Make sure to have the build dependencies available:

    % sudo apt install gcc libc6-dev golang-go

Then either get the repository via:

    % go build -o haproxy-healthcheck github.com/synprosolutions/haproxy-healthcheck

... or to build it from inside the haproxy-healthcheck.git checkout:

    % go build -o haproxy-healthcheck ./main.go

## System integration

To integrate `haproxy-healthcheck`, set up an according systemd service unit:

    % cat /etc/systemd/system/haproxy-healthcheck.service
    [Unit]
    Description = HAProxy healthcheck

    [Service]
    ExecStart = /usr/bin/haproxy-healthcheck
    Restart   = always

Set up the corresponding systemd socket unit:

    % cat /etc/systemd/system/haproxy-healthcheck.socket
    [Socket]
    ListenStream = 16000
    BindIPv6Only = both

    [Install]
    WantedBy = sockets.target

Finally enable the systemd socket activation:

    % sudo systemctl daemon-reload
    % sudo systemctl start haproxy-healthcheck.socket

A sample [haproxy configuration file](https://github.com/SynProSolutions/haproxy-healthcheck/blob/master/haproxy/haproxy.cfg) which might serve as inspiration or for testing is available.

## Author

Michael Prokop, SynProSolutions GmbH

## License

haproxy-healthcheck is licensed under the MIT License.
See [LICENSE](https://github.com/SynProSolutions/haproxy-healthcheck/blob/master/LICENSE) for the full license text.
