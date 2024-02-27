<img src="http://cdn2-cloud66-com.s3.amazonaws.com/images/oss-sponsorship.png" width=150/>

# Trackman

Trackman is a command line tool and Go library that runs multiple commands in a workflow. It support parallel steps, step dependencies, async steps and success checkers.

## Install

Head to [Trackman's releases](https://github.com/cloud66-oss/trackman/releases/latest) and install download the executable for your OS / Architecture. Your version will be updated to the latest version after the first run, so don't worry about the version you pickup first.

## Use

Using Trackman is simple. It uses a YAML file to describe the steps to run and runs them. Here is an example `workflow.yml` file:

```yaml
version: 1
steps:
  - name: list
    command: ls -la
  - name: env
    command: echo $USER
```

Save this file as `workflow.yml` and run it:

```bash
$ trackman run -f workflow.yml
```

Alternatively you can pipe in the workflow into Trackman:

```bash
$ cat workflow.yml | trackman apply -f -
```

This will run both steps in parallel.

### Dependency

Steps can be made dependent to each other:

```yaml
version: 1
steps:
  - name: list
    command: ls -la
  - name: env
    command: echo $USER
    depends_on:
      - list
```

In the example above, `env` will only run once `list` is finished successfully.

You can make a step dependent on more than one step. Such step will only run once all of the dependee steps have finished successfully.

### Success and Failure

By default a step is considered successfully finished when it's done with an exit status of 0.

Sometimes however, there are tasks that run asynchronously and return with 0 immediately but their success will be known later. For example when `kubectl` applies a new configuration to a cluster, its success cannot be determined by the exit status. Trackman supports this by running **probes**.

```yaml
version: 1
steps:
  - name: deploy
    command: kubectl apply -f manifest.yml
    probe:
      command: kubectl wait --for=condition=complete job/myjob
```

This workflow will run `kubectl apply -f manifest.yml` first. If it returns with exit status 0 (it ran successfully), will then run `kubectl wait --for=condition=complete job/myjob` until it returns with exit status 0 and considers the step successful.

Trackman can continue running if a step fails if the step has a `continue_on_fail: true`.

### Timeouts

By default Trackman waits for 10 seconds for each step to complete. If the step fails to complete within 10 seconds, it will consider it failed. This is the same for probes: all probes should return within 10 seconds.

You can change the timeout per step using the `timeout` attribute:

```yaml
version: 1
steps:
  - name: dopy
    command: sleep 60
    timeout: 30s
```

Probes share their step's timeout.

### Metadata

You can add metadata to the workflow file as well as each step. Metadata can be used in step arguments.

```yaml
version: 1
metadata:
    foo: bar
steps:
  - name: list
    metadata:
      fuzz: buzz
    command: ls -la
```

You can use the metadata as arguments of a step:

```yaml
  - name: dump
    metadata:
      foo: bar
    command: "echo {{ index .Metadata \"foo\" }}"
```

Trackman can use Golang template language.

`Metadata` is an attribute on both Step and the entire workflow. You can use `MergedMetadata` instead of `Metadata` to gain access to a merged list of meta data from the step and the workflow. If any value is defined in both places, step will override workflow.

### Work directory

To set the working directory of a step, use `workdir` attribute on a step.

### Environment Variables

All environment variables in commands and their arguments are replaced with `$` values. For example `$HOME` will be replaced with the right home directory address. This is the same for all environment variables available to Trackman at the time it starts.

All environment variables available to Trackman when it starts will be passed on to the step commands.

To specify environment variables that are applies only to a single step, use the `env` attribute:

```yaml
  - name: dump
    env: ["FOO=BAR"]
```

If the assigned environment variable already exists, it will overwrite the OS environment variable for this step.

### Preflight Checks

You can run some checks before the workflow starts. These could be checking for certain binaries or packages to be installed on the machine before the workflow starts.
Each step can have one or more preflight checks. The workflow will run all preflight checks before starting to run the steps. If any of the preflight checks fail, the workflow will not start.

Success of a preflight check is based on the exit status. You can also assign an optional friendly message to each preflight check to be displayed in the event of a failure.

```yaml
steps:
  - name: list
    command: ls -la
    preflights:
      - command: true
        message: "Oh nose!"
```

## Workflow Attributes

The following attributes can be set for the workflow:

| Attribute  | Description  | Default  |
|---|---|---|
| version  | Workflow format version | `1` |
| version  | Any metadata for the workflow | None |
| steps  | List of all workflow steps (See below) | [] |
| logger | Workflow Logger | Default Logger (see below) |
| SessionID | Auto generated 8 digit value for each run of the workflow | Generated |

## Step Attributes

The following attributes can be set for each step:

| Attribute  | Description  | Default  |
|---|---|---|
| metadata  | Any metadata for the step  | None |
| name  | Given name for the step  | `''` |
| command  | Command to run, including arguments  | `''` |
| continue_on_fail  | Continue to the next step even after failure  | `false` |
| timeout  | Timeout after which the step will be stopped. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".   | Never |
| workdir  | Work directory for the step | None |
| probe  | Health probe definition. See above | None |
| depends_on  | List of the steps this one depends on (should run after all of them have successfully finished) | [] |
| preflights  | List of pre-flight checks (see above) | None |
| ask_to_proceed  | Stops the execution of the workflow and asks the user for a confirmation to continue | `false` |
| show_command  | Shows the command and arguments for this step before running it | `false` |
| disabled | Disables the step (doesn't run it). This can be used for debugging or other selective workflow manipulations | `false` |
| env | Environment variables specific to this step | [] |
| logger | Step logger | Workflow logger (see below) |
| SessionID | Auto generated 8 digit value for each run of the workflow | Same as Workflow |

## Trackman CLI

### Global Options

The CLI supports the following global options:

| Option  | Description  | Default  |
|---|---|---|
| config  | Config file | $HOME/.trackman.yaml |
| log-level  | Log level | `info` |
| log-type  | Log Type. Valid options are `stdout`, `stderr`, `discard`, `file` | `stdout` |
| log-format  | Log Format. Valid options are `text` and `json` | `text` |
| log-file  | Log File. If `file` is used as `log-type` then this is used as the filename (path can be included) |  |
| no-update  | Don't update trackman CLI automatically | `false` |

### Run

Runs the given workflow. Use `--help` for more details.

```bash
$ trackman run -f file.yml

$ trackman run -f file.yml -m key1=value -m key2=value
```

### Params

Run command supports the following options

| Option  | Description  | Default  |
|---|---|---|
| file, f  | Workflow file | None |
| timeout | Timeout after which the step will be stopped. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". | 10 seconds |
| concurrency  | Number of concurrent steps to run | Number of CPUs - 1 |
| yes, y  | Answer Yes to all `ask_to_proceed` questions | false |
| metadata, m  | Inline global metadata | None |

### Logging

By default, trackman logs all output to `stdout` and at the `info` level. All logs from all steps are also combined and shown together as they are produced.

You can specify log configuration at the workflow level or for each individual step. If a step has no specific log configuration, it will inherit the configuration of the workflow. Preflight and Probes use the same log configuration as their step.

Log configuration can be defined with the following options:

| Option  | Description  | Default  |
|---|---|---|
| type  | Logger Type. Valid options are `stdout`, `stderr`, `discard` and `file` | `stdout` |
| level  | Log level. Valid values are `error`, `warn`, `info` and `debug` | `info` |
| format  | Log Format. Valid options are `text` and `json` | `text` |
| destination  | Log file (can include path). If type is `file` this is used as the file name. If no path is provided, the current directory is used. | |

Here is an example:

```yaml
version: 1
logger:
  type: "file"
  format: "json"
  destination: "workflow.json"
steps:
  - name: step1
    logger:
      type: "stdout"
  - name: step2
```

In the example above, the workflow and step2 share a file called `workflow.json` for logging. Step1 however will log text to `stdout`.

You can use any attribute from Workflow and Step in naming your log file. Here is an example:

```yaml
version: 1
logger:
  type: "file"
  format: "json"
  destination: "logs/{{ if .Step }}{{.Step.Name}}{{ else }}workflow{{ end }}.json"
```

The example above, will use `workflow.json` for workflow logs but a file named after the step name for each step. You can use Golang templates for this feature. The template is rendered with a context of `Workflow` and `Step` (in the example above, `.Step` is used)

Another example is to use the Workflow SessionID as log file name:

```yaml
version: 1
logger:
  type: "file"
  destination: "logs/{{.Workflow.SessionID}}.log"
```

### Parse

You can use the `parse` command to see how the workflow input yaml file is parsed and what the placeholders (like environment variables) are replaced with before running them. Use `parse` like `run` but without any `timeout` or `concurrency` options:

```bash
$ trackman parse -f workflow.yml
```

### Update

Manually checks for updates. It can also switch the current release channel.

```bash
$ trackman update [--channel name]
```

### Version

Shows the channel and the version

```bash
$ trackman version
```

### Help

Shows help.

```bash
$ trackman help
```

## Update

Trackman updates automatically to the latest available version after each run (except for the `version` command). By default it runs the **stable** channel but you can switch the channel:

```bash
$ trackman update --channel dev
```

This will switch trackman to the **dev** (development) channel and will update it to the latest version of that channel after each run. You can check for updates manually using the `update` command as well. **dev** channel doesn't get automatically updated.

## Release

### Automatic Release

All commits into `master` are built, tested and released on the `edge` channel.
All tags are build, tested and released on the `stable` channel and the binaries are automatically uploaded to Github.
`dev` branch is not automatically built or released.

### Manual Release

If you want to release a new version of Trackman manually, follow these steps:

1. Start a new Release in git flow. Make sure the release name is a valid SemVer text like `1.0.0-rc1` or `2.0.4`.
2. Run `./build.sh CHANNEL`, replacing `CHANNEL` with `dev` or `stable` or `edge`
3. Run `./publish.sh`. This will upload the compiled binaries (previous step) to s3.
4. Create a Github release from the tag and upload the binaries to it.

The last step assumes you have a configured **AWS CLI** installed on your machine with the right permissions to push to `downloads.cloud66.com` bucket.

### Rollback

### Automatic Rollback

On Builtkite, run the needed release again with `FORCE` environment variable equal to `force`.

### Manual Rollback

If you need to rollback a release, switch to the right tag and repeat the build / publish steps but run the build step with the `--force` flag:

```bash
$ ./build.sh CHANNEL --force
```

Having the force flag on a release will force all clients to update all the time which is not desired. You will need to take this flag off when the issue is resolved and push a new version out with a higher version.
