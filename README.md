# yn

`yn` (yaml navigator) is a command-line utility tool crafted for exploration of extensive and multi-documented yaml data. It navigates and highlights the part of the data you are seeking, making it easier to traverse through large data in yaml format. It offers autocompletion and suggestions for field paths.

It is mainly designed to simplify the process of working with Kubernetes manifest yaml files.

![terminal](/images/terminal.png)

## Installation

```bash
go install github.com/semihbkgr/yn@latest
```

## Usage

The data is supplied through `stdin`.

```bash
$ cat data.yaml | yn
```

type the path to traverse `.person.address.city`

![example](/images/example.png)

makes it easy to explore multi-doc outputs of helm templates.

```bash
$ helm template ingress-nginx ingress-nginx --repo https://kubernetes.github.io/ingress-nginx --namespace ingress-nginx | yn
```

![multidoc](/images/multidoc.png)
