*Craig's test docs boilerplate adapted for this project 2022-05-07*

# Tests

Everyone has different definitions of testing terminology. To be clear, i document this here so that you know how this
software refers to different types of test, not because i want to convince you this is the only valid interpretation.

## Ground Rules

* All testing must be fast
    * expect unit tests running slower than 250 tests per second to trigger a CI pipeline build failure. Aim for >1k/sec
      on 2018+ hardware
    * an integration test should usually complete in <1 second (excl. setup/teardown time)
    * Aim for <1 minute per 1k lines of application code (use parallelism). E.g. a 10kLOC app should finish *all*
      testing in under 10 mins.
* No test may depend on other tests and defined ordering of test execution is not permitted
* You should assume all your tests will be run in parallel so be careful to write them appropriately

## Unit Tests

* Unit tests are in `xxx_test.go` files co-located with their respective code in the `/{internal,pkg}/{app,pkg}/**`
  directories
* Expect to have more lines of unit test code than application code
* Unit tests may freely test any arrangement of code in the system, i.e. prefer "classicist" style over "mockist" style
    * NB: unit tests don't have to test only their respective `xxx.go` file's contents in isolation; freely compose
      system parts in-memory and test for desired use case behaviour rather than brittle testing of isolated components
* Unit tests must not depend on anything outside the test process's address space (no live date/time, no files, no
  databases, no network ports, etc. etc.)
* Use the cheapest kind of test-double you can. A dummy is free, stubs are cheap. Favour small, reusable fakes over
  spies and mocks

## Integration Tests

* Prefer unit tests over integration tests, especially for negative condition (sad path) testing
* Integration tests are under `/test/integration/**` directories, they are not co-located with app code
* Each integration test should usually only test 1 integration point
    * E.g. don't have 1 test that asserts about a request to an http endpoint that also touches the DB, that's a slow
      strategy. Instead break
      this into 2 tests that run in parallel. One that asserts about a request to an http endpoint and exercises all
      code right through to an in-memory fake DB rather than a real external DB. Another test that fakes the network
      port but assembles the rest of the system right through to the real DB. Run these tests in parallel.

## End to End (e2e) Tests

* Prefer integration tests over these
* Log screenshots on failures, don't make a dev waste time re-running the test just to see what the error was

## Other Tests

Having read the above, you might still be wondering where another type of test lives. The following types of tests are
not used in this project (yet).

* Security (inc. fuzz tests)
* Operability (observability, tracing & metrics)
* Contract (client-provided tests that must pass before a release can go live)
* Performance
