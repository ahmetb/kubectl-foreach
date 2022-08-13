# kubectl allctx

Run a `kubectl` command in one or more contexts (cluster) based on exact name
match or pattern.

## Usage

```text
Usage:
    kubectl allctx [OPTIONS] [PATTERN]... -- [KUBECTL_ARGS...]

Patterns can be used to match contexts in kubeconfig:
      (empty): matches all contexts
      PATTERN: matches context with exact name
    /PATTERN/: matches context with regular expression
     ^PATTERN: removes results from matched contexts
    
Options:
    -c=NUM       Limit parallel executions (default: 0, unlimited)
    -h/--help    Print help
    -I=VAL       Replace VAL occuring in KUBECTL_ARGS with context name
```

## Examples

**Match to contexts by name:** Run a command ("kubectl version") on contexts `c1`, `c2`
and `c3`:

```sh
kubectl allctx c1 c2 c3 -- version
```

**Match to contexts by pattern:** Run a command on contexts starting with `gke`
and `aws` (regular expression syntax):

```sh
kubectl allctx /^gke/ -- get pods
```

**Match all contexts:** empty context matches all contexts.

```sh
kubectl allctx -- version
```

**Excluding contexts:** Use the matching syntaxes with a `^` prefix to use them
for exclusion. If no matching contexts are specified.

e.g. match all contexts **except** `c1` and except those ending
with `prod` (single quotes for escaping `$` in the shell):

```shell
kubectl allctx ^c1 ^/prod'$'/ -- version
```

**Using with kubectl plugins:** Customize how context name is passed to the command
(useful for kubectl plugins as `--context` must be specified after plugin name).

In this example, `_` is replaced with the context name when calling "kubectl
my_plugin".

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

