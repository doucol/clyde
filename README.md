# clyde

[Project Calico](https://projectcalico.org) Observability Tools

This CLI/TUI application currently allows you to view & watch calico network
flows in near real-time.

The first page you will see is a simple summary totals grouped by SRC namespace
and name, DST namespace & name, and protocol:port.

When on the "home" page (Calico Flow Summary Totals), you can press `r` to
see the flow summary rates (packets/bytes per second). Press `t` to go back
to the home summary totals page.

You can move through the rows with standard vim oriented keystrokes
(up: `k`, down: `j`, top: `g`, bottom: `G`, and arrow / page keys)

Sorting: when in the summary totals page you can sort by SRC namespace & name by
pressing the `n` key. Press it again to reverse the sort. When in the summary
rates page you can do the same but you also have the ability to sort by rates.
`p` for source packets/sec and `P` for destination packets/sec. You can do the same
for the bytes/sec using `b` and `B` respectively. Again, pressing the same key
again will reverse the sort.

Dive into details by hitting \<enter\> on rows and the \<escape\> to back out.

To enable filtering, use the `/` key to show the filter attributes.

## Dependencies

make, go 1.23+, and a Calico OSS 3.30+ cluster

## Quick Start

```bash
make build

# To see the help
bin/clyde --help

# To run the TUI
bin/clyde
```

You can also use the `bin/calico-on-kind` script to quickly create a
[Kind](https://kind.sigs.k8s.io/) based Kubernetes cluster with Calico OSS installed

```bash
# To see help
./bin/calico-on-kind

# Example:
# To create a new cluster, install the GCP demo app, and a set of
# zero trust policies.
VERSION=v3.30 DP=BPF DEMOAPP=true ./bin/calico-on-kind new
```

> **_NOTE:_** ARM based machines may have issues to work through with the DEMOAPP,
> since the GCP demo app does not arm based images published at the time of
> this writing.
