#  mingo

A minimal WebApp + API + DB template/cookie-cutter/starter app, written in Go.

##  Background

Whenever I need a tiny WebApp my go-to "stack" today is Python + venv + Flask + sqlite. It's always worked out well: super-fast to make, easy to change in future and it's cross-platform. Deploying with dependencies has always niggled me with this stack though, I've had a few opinions of the "best" approach to this over the years: cx_freeze, pip, poetry, docker.

##  What Does This Project Do?

Could i use Go to make a "cookie cutter" starter template that's even easier to deploy than a single .py script? I've never written Go before but i already know the answer to that is YES. Static binaries with go:embed FTW! However, the point of this repo is to learn what happens next...

What if the app now needs to scrape a webpage? What if it needs to SSH into another server and grab some files? What if it needs to traverse some directories while reading and writing some files? These are all utterly trivial to do in a Python script. How approachable are these tasks in Go, while still maintaining a decent developer experience & easy deployment strategy?

##  Developer Setup Log

I want to document everything i did to make this, so that i can refresh quickly when i come back to it in future.

I installed:

- Go, i'm using v1.18 (latest available today)
- An IDE or Editor<sup id="b1">[1](#f1)</sup>

###  Initial Setup

I'm winging this with the help of the [Learn Go with Tests site](https://quii.gitbook.io/learn-go-with-tests/) (which has been awesome so far) and the [official doc's tutorial section](https://go.dev/doc/).

Looks like go has historically struggled with modules, however it seems fine today. To create a new app, i needed to:

1. `mkdir ~/Code/local/myproject`
1. `go mod init example.com/myproject`

To refer to a local dependency:

1. `go mod edit -replace example.com/otherproject=../otherproject`
1. `go mod tidy`


###  Basic Developer Experience Tooling

Go itself ships with `go vet` for linting but cli tools like `golangci-lint` can be `brew install`'ed or natively installed. For example, to natively install staticcheck, another lint tool:

1. `go install honnef.co/go/tools/cmd/staticcheck@latest`
1. `staticcheck .`

I went with `golangci-lint` and to configure it in VSCode, i had to add the following to my user settings json:

```
  ...
  "go.lintTool": "golangci-lint",
  "go.lintFlags": [
    "--fast"
  ],
  ...
```

###  The Plan

I want to achieve:

* A simple to hack code base: I want a single Go source file _if_ the code can be kept small and structured in an accessible way
* Easy deployment: I want a single binary file for distribution. Also, sqlite already gives me easy backups & state management (the db is just a file)
* Learn if i think i can be as productive in Go as in Python: I'll test-drive a few common scenarios

So I plan to implement the following and tag the git repo when i'm happy with each step:

1. A basic unit test, i'll implement & test command line flag parsing
   * I'll explore continuous test running after each change in my editor
1. Serve a static net/http endpoint _(I've decided i want minimal dependencies in this project, so i'll use net/http from go's stdlib)_
1. Serve a file endpoint _(I've decided to test-drive [htmx](https://htmx.org/) in this project so i'll serve that .js library)_
   * I'll explore go:embed with the aim of being able to distribute a single binary, cross-compiled for each deployment platform
1. Serve an htmx-driven CRUD page
   * I'll explore live-reloading options
1. Persist data in a sqlite DB
1. Setup schema migrations
1. Connect to a remote json API and parse the results
1. Traverse the filesystem and edit some files
1. (Safely) Run an external command with user provided input

---

<b id="f1">1</b> I've chosen VSCode with the "gopls" language server. Here's a little cheat-sheet for the essential shortcuts to get stuff done: [↩](#b1)

- Lookup:
  - Definition: Opt+F12
  - References: Shift+F12
  - Implementations: Shift+Cmd+F12
  - Parameter docstrings: Shift+Cmd+Space
  - Autocomplete: Ctrl+Space
- Go:
  - Definition: F12
  - Test: Ctrl+Shift+T (custom mapping)
  - File: Cmd+P
  - Back: Ctrl+-
  - Forward: Ctrl+Shift+-
  - Bracket: Shift+Cmd+\
  - Position: Alt+/+<char> (with MetaJump Plugin)
- Refactor:
  - Rename: F2
  - Extract Variable: Alt+Cmd+V (custom mapping)
  - Extract Function: Alt+Cmd+E (custom mapping)
    - Use Ctrl+Shift+Arrow to select code first
  - Inline: ???
  - Move: ???
- Run:
  - App: Ctrl+F5
  - All Tests: Cmd+; A
  - Current Test: Cmd+; C
  - To Breakpoint (Debug): F5
- Close
  - Sidebar: Ctrl+B
  - Bottom bar: Cmd+J
  - Editor Tab: Cmd+W
  - All Editor Tabs: Cmd+K, W
  - Project: Cmd+K, Cmd+F
- Split Editor
  - Vertically: Cmd+\
  - Horizontally: Cmd+K, Cmd+\
- Focus Editor: Escape (custom mapping)
- Toggle Focus Between Terminal & Editor: Ctrl+` (custom mapping)
- File Explorer: Shift+Cmd+E
- Find: Shift+Cmd+F
- VCS: Ctrl+Shift+G
- Undiscoverable-but-Useful Snippets:
  - helloweb: Hello world webapp
  - pkgm: Package main + main func
  - ims: Import packages
  - cos: Define constants
  - tf: Test function
  - ef: Example function
  - bf: Benchmark function
  - tdt: Table driven test
  - ff: fmt.Printf("...", var)
  - lf: Printf with var + newline
  - lv: Log variable content
  - las: Http listen and serve
  - hf: Http Handle Func
  - wr: Http response writer
  - hand: Http handler
  - rd: Http redirect
  - herr: Http error
  - df: defer statement
  - meth: Method declaration
  - sort: Custom sort impl
