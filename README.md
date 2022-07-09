# kubectl allctx

Run a `kubectl` command in one or more contexts (cluster) based on exact name
match or pattern.

Example:

```sh
# run "kubectl version" on all contexts starting with 'gke'
# note that we're escaping the ~ character, which the shell would otherwise expand
kubectl allctx '~^gke' -- version

# run a command on multiple contexts (exact name match)
kubectl allctx ctx1 ctx2 -- get pods

# run on all context, but two at a time (empty regexp matches all)
kubectl allctx -c 2 '~' -- version

# specify where the context name gets passed in the command executed (_ replaced with context name)
kubectl allctx -I _ '~test$' -- my_plugin -ctx=_

# Bypass prompt with using '-q' flag
kubectl allctx -q '~' --auto get nodes -o wide
```

## Install

TODO(ahmetb): Add krew installation method

```
go install github.com/ahmetb/kubectl-allctx@latest
```
