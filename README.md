# kubectl allctx

Run a `kubectl` command in one or more contexts (cluster) based on exact name
match or pattern.

## Usage

**Exact match:** Run a command ("kubectl version") on contexts `c1`, `c2`
and `c3`:

```sh
kubectl allctx c1 c2 c3 -- version
```

**All contexts:** empty context matches all contexts.

```sh
kubectl allctx -- version
```

**Pattern matching:** Run a command on contexts starting with `gke` and `aws` (regular
expression syntax):

```sh
kubectl allctx /^gke/ -- get pods
```

**Mixing patterns and exact matches:** Matched contexts are added together.

```sh
kubectl allctx c1 c2 /re1/ /re2 -- version
```

**Exclusion:**  Run on all contexts except `c1` and ending with `prod`:

```shell
kubectl allctx -c1 -/prod$/ -- version
```

**Argument customization:** Customize how context name is passed to the command
(useful for kubectl plugins as `--context` must be specified after plugin name).
In this example, `_` is replaced with the context name.

```shell
kubectl allctx -I_ -- my_plugin -ctx=_
```

**Limit parallelization:** Only run 3 commands at a time:

```
kubectl allctx -c 3 /^gke-/
```

## Install

Currently, the `go` command is the only way to install
(make sure `~/go/bin` is in your `PATH`):

```
go install github.com/ahmetb/kubectl-allctx@latest
```

