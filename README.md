## XFSGO

Official Go implementation of the XFS protocol.

### Usage

To learn more about the available `xfsgo` commands, use
`xfsgo` help or type a command followed by the `-h` flag

```bash
xfsgo -h
```

For further details on options, see the [documents](https://docs.xfs.tech/)

### Building From Source

To build from source, Go 1.13 or above must be
installed on the system. Clone the repo and run
make:

```bash
git clone https://github.com/xfs-network/xfsgo.git
cd xfsgo && make
```

### Contribute

If you'd like to contribute to xfsgo please fork, fix, commit and
send a pull request. Commits who do not comply with the coding standards
are ignored (use gofmt!). If you send pull requests make absolute sure that you
commit on the `develop` branch and that you do not merge to master.
Commits that are directly based on master are simply ignored.

### License

XFS is released under the open source [MIT](./LICENSE) license and is offered “AS IS”
without warranty of any kind, express or implied. Any security provided
by the XFS software depends in part on how it is used, configured, and
deployed. XFS is built upon many third-party libraries, xfs.tech makes
no representation or guarantee that XFS or any third-party libraries
will perform as intended or will be free of errors, bugs or faulty
code. Both may fail in large or small ways that could completely or
partially limit functionality or compromise computer systems. If you use
or implement XFS, you do so at your own risk. In no event will
xfs.tech be liable to any party for any damages whatsoever, even if it
had been advised of the possibility of damage.
