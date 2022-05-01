#  mingo

A minimal WebApp + API + DB template/cookie-cutter/starter app, written in Go.

##  Background

Whenever I need a tiny WebApp my go-to "stack" today is Python + venv + Flask + sqlite. It's always worked out well: super-fast to make, easy to change in future and it's cross-platform. Deploying with dependencies has always niggled me with this stack though, I've had a few opinions of the "best" approach to this over the years: cx_freeze, pip, poetry, docker.

##  What Does This Project Do?

Could i use Go to make a "cookie cutter" starter template that's even easier to deploy than a single .py script? I've never written Go before but i already know the answer to that is YES. Static binaries with go:embed FTW! However, the point of this repo is to learn what happens next...

What if the app now needs to scrape a webpage? What if it needs to SSH into another server and grab some files? What if it needs to traverse some directories while reading and writing some files? These are all utterly trivial to do in a Python script. How approachable are these tasks in Go, while still maintaining a decent developer experience & easy deployment strategy?

##  Developer Setup

I want to document everything i did to make this, so that i can refresh quickly when i come back to it in future.

I installed:

- Go, i'm using v1.18 (latest available today)
- An IDE or Editor<sup id="b1">[1](#f1)</sup> VSCode with gopls in my case

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

### Development Log

Notes of my experiences as I implement each part of the plan above.

#### Milestone 1: Testing & Command Line Args

The built in flag package can be made to comply with GNU-ish conventions but it's not the default. Probably the Go defaults reflect Plan9 OS's conventions? For example, i'd want single hyphen with single letter flags and double hyphen with word flags, e.g. "-h" and "--help" but flag wants to steer you to "-help". Indeed with my program, all 3 have become valid since i can't stop handling of the "-help" case. Making flag usage testable was ok-ish to do.

* Feeling: stoked, i've been wanting to test-drive Go for ages and it's fun
* Time spent (estimate): 2 hours

####  Milestone 2: Serve an HTTP Endpoint

You could do this *much* faster if you're willing to go with really basic behaviours but i wanted formatted logging, clean termination, parse-don't-validate style modelling of constraints like "valid network port" in the type system and of course i wanted to keep everything testable, which mostly means i wanted to use the dependency injection pattern everywhere.

Mirroring my experience yesterday with the flag package, i found log to be a little bit under-powered but it didn't stop me doing what i wanted and ultimately that's what counts. I'm absolutely conscious that i could have had an easier ride with config, flags and logging by installing dependencies like viper, cobra & golog but i want to stick with zero external deps.

Finding https://gist.github.com/creack/4c00ee404f2d7bd5983382cc93af5147 was a *major* help. I had a bunch of "basic" behaviours i wanted for my server and i'd have implemented them but this person did it nicer than i know my first attempt in Go would have been! They even added tracing which was a bonus that wasn't on my must-have's list. One thing they were missing was logging of HTTP Status code - this meant i had to ditch their defered goroutine approach because i need strict ordering (http handler *then* log result) but i have a sneaky suspicion there's a performance change by doing this. I don't know if it's positive or negative so i'll need to load test it later when i do my acceptance testing at the end.

I'm starting to form an opinion of my productivity in Go, bearing in mind i've never written Go before, i'm more productive right now than in vanilla Java (NB: excluding use of Spring). I'm more productive sooner than i expected, i attribute that to Go reliably doing what i expect. i haven't had a "Wow! I didn't expect that?!" moment to shake my confidence yet. There are some oddities right enough: date/time formatting literals, although it works just fine; also the const keyword is less useful than i'd hope for.

I'm a little bit unsure about some loose ends such as the HTTP Content-type header i'm not setting anywhere yet, i suspect it's probably serving a default text/html type for me. I haven't told the ServeMux router what HTTP methods to accept. I have no idea how to apply backpressure to incoming requests. I'm a little bit worried about unneccessary heap pressure in the hot-path of serving requests, for example, LoggingResponseWriter is created for every request - now it's ultimately just an int and a reference to an existing writer, but in the hot-path, you don't want to be throwing fresh junk for the garbage collector each time. That said, i'm serving requests in around 17 micro seconds on an M1 Macbook Air, that'd be 50k-ish req/sec roughly.

I love the [Learn Go With Tests](https://quii.gitbook.io/learn-go-with-tests/meta/why) site, as i've gone through more of it, i've just found more that i like.

I think my initial goal of a single .go source file is misguided. That idea was based on my thoughts & desires for simplicity before i had experienced writing Go. Now with some experience, a single .go source file is not the right way to go about that.

I like the built in testing facilities, there's nice touches like you can provide "example" tests and they're included in generated docs.

Another thing i like about Go is the community has definitely got the memo on unit testing, this is a breath of fresh air when i'm coming from working with Java where developers usually lean heavily towards integration testing. There's a place for both of course but many Java developers would disagree with unit testing being less painful because the common Java unit testing idiom is to treat "unit test" as meaning test a class at a time which is of course painful.

* Feeling: reflective, Python is more expressive but so far i prefer Go over Java8
* Time spent (estimate): 8 hours


####  Milestone 3: Serve a File Endpoint

Ok now THAT was cool. Go embed is *awesome*. I can see so many use cases for simplifying distribution. My entire binary is now just over 6mb (only 40k of that is an embedded .js file). I *wish* my usual docker containers were only 6mb, this really is the neatest feature of Go so far. It was so ridiculously easy too.

I'm not sure about my use of Go's testing facilities. I feel like my tests are quite long winded.

I took a detour today and in a separate toy project i implemented a test that `os.Exec()`'s a "go run ." in another process and i got that working so if i change my mind about the current setup in this project (I call `main()` in a go routine) then i can switch over.

* Feeling: vindicated investing time test-driving Go, go:embed is what i hoped would be possible in Go
* Time spent (estimate): 15 minutes

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
