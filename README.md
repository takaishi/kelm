# Kelm

[![Go Report Card](https://goreportcard.com/badge/github.com/takaishi/kelm)](https://goreportcard.com/report/github.com/takaishi/kelm)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[license]: https://github.com/takaishi/kelm/blob/master/LICENSE

Interactive kubernetes operator like a peco and Emacs helm.

![](./docs/images/example.gif)

## Install

```
$ brew tap takaishi/homebrew-fomulas
$ brew install kelm
```

## Usage

```
$ kelm
```

##  Custom Action

You can write custom action to `~/.kelm`.

For example:

```yaml
---
actions:
  pods:
    - name: "log"
      command:  "kubectl -n {{ .Namespace }} log {{ .Obj.metadata.name }}"
  nodes:
    - name: "ssh"
      variables:
        - name: address
          jsonpath: '{.status.addresses[?(@.type=="InternalIP")].address}'
      command: 'ssh {{ .address }}'

```

## Author

[r_takaishi](https://github.com/takaishi)