# `token-injector` Tool

The `token-injector` tool can get Google Cloud ID token when running under GCP Service Account (for example, GKE Pod with Workload Identity).

## `token-injector` Command Syntax

```text
NAME:
   token-injector - generate ID token with current Google Cloud service account

USAGE:
   token-injector [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --refresh      auto refresh ID token before it expires (default: true)
   --file value   write ID token into file (stdout, if not specified)
   --help, -h     show help (default: false)
   --version, -v  print the version
```
