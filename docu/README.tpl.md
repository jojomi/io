# io

[![Godoc Reference](https://godoc.org/github.com/jojomi/io?status.svg)](http://godoc.org/github.com/jojomi/io)
![Go Version](https://img.shields.io/github/go-mod/go-version/jojomi/io)
![Last Commit](https://img.shields.io/github/last-commit/jojomi/io)
[![Go Report Card](https://goreportcard.com/badge/jojomi/io)](https://goreportcard.com/report/jojomi/io)
[![License](https://img.shields.io/badge/License-MIT-orange.svg)](https://github.com/jojomi/io/blob/master/LICENSE)

Take data, make documents!

## Overview

![io overview](docu/overview.svg)

`io` is supposed to be a small and useful tool for reworking data from JSON, YAML, or CSV sources into any text or HTML format.

The templates used for the transformation feature all the elements of [Go Templates](https://pkg.go.dev/text/template)
plus a set of useful [functions](#template-functions).

Gems are the `exec` functions from `tplfuncs` that, combined with the line based matchers and filters,
can be used to create dynamic auto-generated documents.

Also this tool can be used to setup simple and easy includes for the system Hostsfile at `/etc/hosts`, see [here](#auto-generate-etchosts).

## How to Use

```
{{ exec "io --help" }}
```

## Example

{{ $input := "test/input/simple.yml" -}}
With input data from [{{ $input }}]({{ $input }})

``` yml
{{ printf "cat %s" $input | exec }}
```

{{ $template := "test/template/creator.html" -}}
and the template [{{ $template }}]({{ $template }})

``` yml
{{ printf "cat %s" $template | exec }}
```

you can use `io` to get this result:

{{ $ioCmd := printf "io -i %s -t %s" $input $template -}}
``` shell
> {{ $ioCmd }}
{{ exec $ioCmd | trim }}
```

If you want to overwrite values from the input data uses `--overwrite` like this:
{{ $ioCmdOverwrite := printf "io --input %s --template %s --overwrite creator.age=62 --overwrite creator.name=Walther" $input $template -}}
``` shell
> {{ $ioCmdOverwrite }}
{{ exec $ioCmdOverwrite | trim }}
```

### Auto-generate /etc/hosts

Create a file named `/etc/hosts.gen`:

``` shell
{{ include "docu/hosts.gen" }}
```

Now you can place an arbitrary number of files in `/etc/hosts.d` (create the directly like this: `mkdir /etc/hosts`).
These files will be inlined as described in the template. The automatic comment makes sure that you always know which source file to edit if there is anything to be changed.

To update the hostsfile the system will use at `/etc/hosts`, run this:

``` shell
{{ regexReplaceAll "^#\\s*|\\s*#$" (include "docu/hosts.gen" | match "sudo io") "" }}
```

Other content around the `range` operation is left untouched, but can still only be edited in `/etc/hosts.gen` otherwise it would be overwritten.


## Template Functions

* all functions defined in [Masterminds/**sprig**](http://masterminds.github.io/sprig/)
* all functions defined in [jojomi/**tplfuncs**](https://github.com/jojomi/tplfuncs) (the `exec*` variants are only avaiable when `--allow-exec` is given when calling `io` due to security implications)

A quick introduction to Golang Templates can be found at [Hugo](https://gohugo.io/templates/introduction).

## How to Install

``` shell
go install {{ regexReplaceAllLiteral "\\.git$" (exec "git config --get remote.origin.url" | trim | replace "git@" "" | replace "https://" "" | replace ":" "/" ) "" -}} @latest
```

## Who uses it?

`io` does [itself](https://en.wikipedia.org/wiki/Eating_your_own_dog_food), see [build.sh](build.sh) which generates this very document from [docu/README.tpl.md](docu/README.tpl.md). It shows how to use `exec` functions as well, but does not take dynamic input data.