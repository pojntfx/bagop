# bagop

Build for all Go-supported platforms by default, disable those which you don't want.

[![hydrun CI](https://github.com/pojntfx/bagop/actions/workflows/hydrun.yaml/badge.svg)](https://github.com/pojntfx/bagop/actions/workflows/hydrun.yaml)
[![Matrix](https://img.shields.io/matrix/bagop:matrix.org)](https://matrix.to/#/#bagop:matrix.org?via=matrix.org)

## Overview

bagop is a simple build tool for Go which tries to build your app for all platforms supported by Go by default. Instead of manually adding specific `GOOS`es and `GOARCH`es, bagop builds for all valid targets by default, and gives you the choice to disable those which you don't want to support or which can't be supported.

## Installation

Static binaries are also available on [GitHub releases](https://github.com/pojntfx/bagop/releases).

On Linux, you can install them like so:

```shell
$ curl -L -o /tmp/bagop "https://github.com/pojntfx/bagop/releases/latest/download/bagop.linux-$(uname -m)"
$ sudo install /tmp/bagop /usr/local/bin
```

On macOS, you can use the following:

```shell
$ curl -L -o /tmp/bagop "https://github.com/pojntfx/bagop/releases/latest/download/bagop.darwin-$(uname -m)"
$ sudo install /tmp/bagop /usr/local/bin
```

On Windows, the following should work (using PowerShell as administrator):

```shell
PS> Invoke-WebRequest https://github.com/pojntfx/bagop/releases/latest/download/bagop.windows-x86_64.exe -OutFile \Windows\System32\bagop.exe
```

You can find binaries for more operating systems and architectures on [GitHub releases](https://github.com/pojntfx/bagop/releases).

## Tutorial

Let's assume we have a Go app called `hello-world` and we want to build it for as many platforms as possible using bagop. This is the `main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
```

We could now start to manually (cross-)compile by running `go build -o hello-world main.go` and setting `GOOS` or `GOARCH`, or use simplified process with bagop. To build `hello-world` for all Go-supported platforms, simply run:

```shell
$ bagop -b hello-world main.go
2021/07/28 14:34:40 building aix/ppc64 (out/hello-world.aix-ppc64)
2021/07/28 14:34:40 building android/386 (out/hello-world.android-i686)
2021/07/28 14:34:41 building android/amd64 (out/hello-world.android-x86_64)
2021/07/28 14:34:41 building android/arm (out/hello-world.android-armv7l)
2021/07/28 14:34:41 could not build for platform android/arm: err=exit status 2, stdout=, stderr=# command-line-arguments
loadinternal: cannot find runtime/cgo
/usr/local/go/pkg/tool/linux_amd64/link: running gcc failed: exit status 1
gcc: error: unrecognized command line option ‘-marm’; did you mean ‘-mabm’?
```

As you can see, we get an error for Android (Android and iOS require additional support). We decide we don't want to support Android (and iOS), so let's re-run the command with these platforms disabled:

```shell
$ bagop -b hello-world -x '(android/*|ios/*|openbsd/mips64)' main.go
2021/07/28 14:36:50 building aix/ppc64 (out/hello-world.aix-ppc64)
2021/07/28 14:36:51 skipping android/386 (platform matched the provided regex)
2021/07/28 14:36:51 skipping android/amd64 (platform matched the provided regex)
2021/07/28 14:36:51 skipping android/arm (platform matched the provided regex)
2021/07/28 14:36:51 skipping android/arm64 (platform matched the provided regex)
2021/07/28 14:36:51 building darwin/amd64 (out/hello-world.darwin-x86_64)
2021/07/28 14:36:51 building darwin/arm64 (out/hello-world.darwin-aarch64)
2021/07/28 14:36:51 building dragonfly/amd64 (out/hello-world.dragonfly-x86_64)
2021/07/28 14:36:51 building freebsd/386 (out/hello-world.freebsd-i686)
2021/07/28 14:36:51 building freebsd/amd64 (out/hello-world.freebsd-x86_64)
2021/07/28 14:36:51 building freebsd/arm (out/hello-world.freebsd-armv7l)
2021/07/28 14:36:52 building freebsd/arm64 (out/hello-world.freebsd-aarch64)
2021/07/28 14:36:52 building illumos/amd64 (out/hello-world.illumos-x86_64)
2021/07/28 14:36:52 skipping ios/amd64 (platform matched the provided regex)
2021/07/28 14:36:52 skipping ios/arm64 (platform matched the provided regex)
2021/07/28 14:36:52 building js/wasm (out/hello-world.js-wasm.wasm)
2021/07/28 14:36:52 building linux/386 (out/hello-world.linux-i686)
2021/07/28 14:36:52 building linux/amd64 (out/hello-world.linux-x86_64)
2021/07/28 14:36:52 building linux/arm (out/hello-world.linux-armv7l)
2021/07/28 14:36:53 building linux/arm64 (out/hello-world.linux-aarch64)
# ...
2021/07/28 14:36:55 building openbsd/arm64 (out/hello-world.openbsd-aarch64)
2021/07/28 14:36:56 skipping openbsd/mips64 (platform matched the provided regex)
2021/07/28 14:36:56 building plan9/386 (out/hello-world.plan9-i686)
2021/07/28 14:36:56 building plan9/amd64 (out/hello-world.plan9-x86_64)
2021/07/28 14:36:56 building plan9/arm (out/hello-world.plan9-armv7l)
2021/07/28 14:36:56 building solaris/amd64 (out/hello-world.solaris-x86_64)
2021/07/28 14:36:56 building windows/386 (out/hello-world.windows-i686.exe)
2021/07/28 14:36:56 building windows/amd64 (out/hello-world.windows-x86_64.exe)
2021/07/28 14:36:56 building windows/arm (out/hello-world.windows-armv7l.exe)
```

If we now check the `out` directory, we can see that we now have successfully built binaries for all supported platforms:

```shell
$ file out/*
out/hello-world.aix-ppc64:          64-bit XCOFF executable or object module
out/hello-world.darwin-aarch64:     Mach-O 64-bit arm64 executable, flags:<|DYLDLINK|PIE>
out/hello-world.darwin-x86_64:      Mach-O 64-bit x86_64 executable
out/hello-world.freebsd-x86_64:     ELF 64-bit LSB executable, x86-64, version 1 (FreeBSD), statically linked, Go BuildID=fV2BDKuHDCvCOj7-m7vv/W9RphXmMqdRiU2gaOpqn/3IGIUAncn0Ru4SxodzxW/c9mup3kmDEeCEdArhSBs, not stripped
out/hello-world.illumos-x86_64:     ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib/amd64/ld.so.1, Go BuildID=Kw7ILaj6FXiE9AjDNTsP/iQzcICQFRvgwBPBBHmyx/X-HuChFgFJmkDLQE17-7/bBmY7Qsz1mOL63EZfHxx, not stripped
out/hello-world.linux-mips64:       ELF 64-bit MSB executable, MIPS, MIPS-III version 1 (SYSV), statically linked, Go BuildID=a89_1vHuZqx8j8tbx6rv/h6exUcnTnEzLeg5zwG2j/8JEsGrUTMIDOL_mdzYYe/bzMrWSV38cJMDdF71rn6, not stripped
out/hello-world.plan9-i686:         Plan 9 executable, Intel 386
out/hello-world.plan9-x86_64:       data
# ...
out/hello-world.solaris-x86_64:     ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib/amd64/ld.so.1, Go BuildID=GjGEjnApRUWlg95uaQOs/Z-0sBhR9hzCdM8RVekoQ/2-IjmJ6MK_M0nMTUXJhC/2cHkBOpKkPh7F6-rgmBa, not stripped
out/hello-world.windows-armv7l.exe: PE32 executable (console) ARMv7 Thumb (stripped to external PDB), for MS Windows
out/hello-world.windows-i686.exe:   PE32 executable (console) Intel 80386 (stripped to external PDB), for MS Windows
out/hello-world.windows-x86_64.exe: PE32+ executable (console) x86-64 (stripped to external PDB), for MS Windows
```

Now, let's add a few compiler flags to make the build binaries fully static; we can do this by using the `-p` flag and manually setting the build command. We'll also set `-j` to allow parallel builds:

```shell
$ CGO_ENABLED=0 bagop -j "$(nproc)" -b hello-world -x '(android/*|ios/*|openbsd/mips64)' -p "go build -a -ldflags '-extldflags \"-static\"' -o \$DST main.go"
2021/07/28 14:36:50 building aix/ppc64 (out/hello-world.aix-ppc64)
2021/07/28 14:36:51 skipping android/386 (platform matched the provided regex)
2021/07/28 14:36:51 skipping android/amd64 (platform matched the provided regex)
2021/07/28 14:36:51 skipping android/arm (platform matched the provided regex)
# ...
```

If we now check the output again, you can see that the binaries are now fully static:

```shell
$  ldd out/hello-world.linux-x86_64
        not a dynamic executable
```

🚀 **That's it!** We've successfully added support for a total of _38_ target platforms to this app.

If you're enjoying bagop, the following projects might also be of help to you too:

- Also want to test these cross-compiled binaries? Check out [hydrun](https://github.com/pojntfx/hydrun)!
- Need to cross-compile CGo? Check out [bagccgop](https://github.com/pojntfx/bagccgop)!
- Want to build fully-featured desktop GUI for all these platforms without CGo? Check out [Lorca](https://github.com/zserge/lorca)!
- Want to use SQLite without CGo? Check out [cznic/sqlite](https://gitlab.com/cznic/sqlite)!

## Reference

```shell
$ bagop --help
Build for all Go-supported platforms by default, disable those which you don't want.

Example usage: bagop -b mybin -x '(android/arm|ios/*|openbsd/mips64)' -j "$(nproc)" 'main.go'
Example usage (with plain flag): bagop -b mybin -x '(android/arm|ios/*|openbsd/mips64)' -j "$(nproc)" -p 'go build -o $DST main.go'

See https://github.com/pojntfx/bagop for more information.

Usage: bagop [OPTION...] '<INPUT>'
  -b, --bin string          Prefix of resulting binary (default "mybin")
  -d, --dist string         Directory build into (default "out")
  -x, --exclude string      Regex of platforms not to build for, i.e. (windows/386|linux/mips64)
  -e, --extra-args string   Extra arguments to pass to the Go compiler
  -g, --goisms              Use Go's conventions (i.e. amd64) instead of uname's conventions (i.e. x86_64)
  -j, --jobs int            Maximum amount of parallel jobs (default 1)
  -p, --plain               Sets GOARCH, GOARCH and DST and leaves the rest up to you (see example usage)
  -v, --verbose             Enable logging of executed commands
```

## Contributing

To contribute, please use the [GitHub flow](https://guides.github.com/introduction/flow/) and follow our [Code of Conduct](./CODE_OF_CONDUCT.md).

To build bagop locally, run:

```shell
$ git clone https://github.com/pojntfx/bagop.git
$ cd bagop
$ go run ./cmd/bagop/main.go --help
```

Have any questions or need help? Chat with us [on Matrix](https://matrix.to/#/#bagop:matrix.org?via=matrix.org)!

## License

bagop (c) 2024 Felicitas Pojtinger and contributors

SPDX-License-Identifier: AGPL-3.0
