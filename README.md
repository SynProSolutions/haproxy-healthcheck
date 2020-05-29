# haproxy-healthcheck

`haproxy-healthcheck` is a tool for usage as [agent-check in haproxy's](https://cbonte.github.io/haproxy-dconv/2.0/configuration.html#5.2-agent-check) server configuration, to report back CPU idle time or the server's administrative state.

The reported integer percentage is similar to what the `vmstat(8)` tool reports in its `id` column when being invoked with a delay of 1 second, like:

     vmstat 1
    procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----
     r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st
     2  0      0 1939504 213764 4155308    0    0    40    64    5    9  9  5 86  0  0
     0  0      0 1942048 213768 4153104    0    0     0    20 2036 5618  7  7 86  0  0
     0  0      0 1941764 213768 4152592    0    0     0    20 2206 5761  7  8 85  0  0
     [...]

The file `/var/run/haproxy-healtcheck` can be used to control the server's administrative state.
If the file exists and contains a support keyword, its content is reported *instead* of the integer percentage.
The following keywords are supported (the optional description strings aren't supported yet):

* down
* drain
* failed
* maint
* stopped
* ready
* up

## Usage

*haproxy-healthcheck* is supposed to be used with [systemd socket activation](https://vincent.bernat.ch/en/blog/2018-systemd-golang-socket-activation).
For testing its integration, you can manually start it to listen on port 2048 via:

    % systemd-socket-activate -l 2048 ./main

## Installation / Build instructions

Make sure to have the build dependencies available:

    % sudo apt install gcc libc6-dev

Then either `go get` the repository:

    % go get https://github.com/SynProSolutions/haproxy-healthcheck

... or to build from inside haproxy-healtcheck.git checkout:

    % go get
    % go build main.go

## Author

Michael Prokop, SynProSolutions

## License

haproxy-healthcheck is licensed under the MIT License.
See [LICENSE](https://github.com/SynProSolutions/haproxy-healthcheck/blob/master/LICENSE) for the full license text.