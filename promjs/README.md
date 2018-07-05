# PromJS - Prometheus in the browser

This is a modification to Prometheus that is maintained as part of the Lightbend
fork. It uses GopherJS to translate into pure Javascript.

It consists of the packages:
* github.com/prometheus/prometheus/promjs - this folder, with `main.go` as the entrypoint for the
  library.
* github.com/prometheus/prometheus/promcache - cache logic.

It also relies on a fork of GopherJs that supports creating an in-memory virtual file system (VFS).
This is necessary as prometheus relies heavily on filesystem operations, and GopherJS by
default does not support these operations.

## GopherJS Fork

We rely on the GopherJS fork/branch at https://github.com/badgerodon/gopherjs/tree/ext.

There are only a few commits on this branch:
https://github.com/gopherjs/gopherjs/compare/master...badgerodon:ext

It is currently vendored in, so no additional work is needed to build.

(@jsravn: Going forward, we should probably fork GopherJS into the Lightbend repo
so we can keep it up to date with upstream.)

# Building

First get gopherjs:

    go get -u github.com/gopherjs/gopherjs

Then from the `promjs` folder:

    gopherjs build -o promjs.js

Check in the resulting files, create a PR and you're done.