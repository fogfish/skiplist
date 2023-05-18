<p align="center">
  <h3 align="center">Skip List</h3>
  <p align="center"><strong>probabilistic, mutable list-based data structure</strong></p>

  <p align="center">
    <!-- Version -->
    <a href="https://github.com/fogfish/skiplist/releases">
      <img src="https://img.shields.io/github/v/tag/fogfish/skiplist?label=version" />
    </a>
    <!-- Documentation -->
    <a href="https://pkg.go.dev/github.com/fogfish/skiplist">
      <img src="https://pkg.go.dev/badge/github.com/fogfish/skiplist" />
    </a>
    <!-- Build Status  -->
    <a href="https://github.com/fogfish/skiplist/actions/">
      <img src="https://github.com/fogfish/skiplist/workflows/test/badge.svg?branch=main" />
    </a>
    <!-- GitHub -->
    <a href="http://github.com/fogfish/skiplist">
      <img src="https://img.shields.io/github/last-commit/fogfish/skiplist.svg" />
    </a>
    <!-- Coverage -->
    <a href="https://coveralls.io/github/fogfish/skiplist?branch=main">
      <img src="https://coveralls.io/repos/github/fogfish/skiplist/badge.svg?branch=main" />
    </a>
    <!-- Go Card -->
    <a href="https://goreportcard.com/report/github.com/fogfish/skiplist">
      <img src="https://goreportcard.com/badge/github.com/fogfish/skiplist" />
    </a>
  </p>
</p>

---

Package `skiplist` implements a probabilistic, mutable list-based data structure that are a simple and efficient substitute for balanced trees. The algorithm is well depicted by [the article](http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.17.524).


## Installing 

The latest version of the library is available at `main` branch. All development, including new features and bug fixes, take place on the `main` branch using forking and pull requests as described in contribution guidelines. The stable version is available via Golang modules.

1. Use `go get` to retrieve the library and add it as dependency to your application.

```bash
go get -u github.com/fogfish/skiplist
```

2. Import it in your code

```go
import (
  "github.com/fogfish/skiplist"
)
```

## Quick Example

Here is a minimal example on creating an instance of the skiplist. See the full [example](examples/skiplist.go)

```go
package main

import (
  "github.com/fogfish/skiplist"
  "github.com/fogfish/skiplist/ord"
)

func main() {
  list := skiplist.New[int, string](ord.Type[int]())

  skiplist.Put(list, 5, "instance")
  skiplist.Get(list, 5)
  skiplist.Remove(list, 5)
}
```

## How To Contribute

The library is [MIT](LICENSE) licensed and accepts contributions via GitHub pull requests:

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request


The build and testing process requires [Go](https://golang.org) latest version.

**Build** and **run** in your development console.

```bash
git clone https://github.com/fogfish/skiplist
cd skiplist
go test
go test -run=^$ -bench=. -cpu 1
```

### commit message

The commit message helps us to write a good release note, speed-up review process. The message should address two question what changed and why. The project follows the template defined by chapter [Contributing to a Project](http://git-scm.com/book/ch5-2.html) of Git book.

### bugs

If you experience any issues with the library, please let us know via [GitHub issues](https://github.com/fogfish/skiplist/issue). We appreciate detailed and accurate reports that help us to identity and replicate the issue. 

## License

[![See LICENSE](https://img.shields.io/github/license/fogfish/skiplist.svg?style=for-the-badge)](LICENSE)