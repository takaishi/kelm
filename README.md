# Kelm

Interactive kubernetes operator like a peco and Emacs helm.

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

```

## Licence

[MIT](https://github.com/takaishi/tool/blob/master/LICENCE)

## Author

[r_takaishi](https://github.com/takaishi)