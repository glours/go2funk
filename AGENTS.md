# Project: go2funk

Purely functional structures for Go (`Option`, `Either`, `Try`, `List`, `Pair`),
inspired by Vavr. Small library, no framework, no runtime dependencies.

A rework is under way: the public types are moving from interfaces to concrete
types so that `Map` can change its type parameter, which Go 1.27 generic methods
finally allow. **Ask before changing the public API** — several of the rules
below encode decisions that are already settled.

## Build & Test

- Build: `go build ./...`
- Test all: `go test ./...` (add `-race` before pushing)
- Test one package: `go test ./api/control/`
- Test one function: `go test ./api/control/ -run TestOptionMap`
- Coverage: `go test -cover ./...`
- Vet: `go vet ./...`
- Benchmarks: `go test -bench . -benchmem ./...`
- Requires **Go 1.27+** (generic methods). Never lower the `go` directive in `go.mod`.

## Zero runtime dependencies — hard rule

The README promises a dependency-free library. This is the project's defining
constraint, not a preference.

- `api/` **must import nothing outside the Go standard library.**
- Verify — this command must print nothing:

  ```sh
  go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./... | grep -v go2funk
  ```

- Test-only dependencies are tolerated but kept to a minimum. Prefer the stdlib
  `testing` package over an assertion library for new tests.
- Never add a dependency to solve something the stdlib already solves
  (`slices`, `maps`, `cmp`, `iter`, `hash/maphash`, `errors`, `sync`).
- If you believe a runtime dependency is unavoidable: **stop and ask.** Do not
  add it and explain afterwards.

## TDD — mandatory

Every behaviour change follows red → green → refactor:

1. **Red** — write a test that fails for the right reason. Run it and show the
   failure. A test that passes on the first run has proven nothing.
2. **Green** — write the minimum production code to make it pass.
3. **Refactor** — clean up with the tests green.

- Never write production code without a failing test first.
- Never delete, skip, or weaken a test to make a build pass. If a test is wrong,
  say so and ask before touching it.
- Bug fixes start with a test that reproduces the bug.
- Do not chase a coverage number. This repo already had 96–100% coverage while
  shipping three critical bugs.

## Test rules

- **Black box first.** New tests go in a `package X_test` file and consume only
  the exported API. Internal tests (`package X`) are allowed only for genuinely
  unexported logic, and must be justified in a comment.
- **Assert values, not booleans.** For error-carrying types, assert the *cause*,
  not just `IsFailure()`. For containers, assert the *contents*, not just the
  length.
- **No magic constants.** Never assert against an opaque literal (a hash, a
  serialized blob) without a comment explaining how it was derived and what it
  guarantees. Prefer asserting the property (`hash(a) != hash(b)`) over the value.
- **Table-driven** for anything with more than two cases, with a named `name`
  field and `t.Run`.
- **No shared mutable state between tests.** Build fixtures inside the test or a
  helper, not in package-level `var` blocks.
- **Test the zero value.** Every exported type must behave correctly when
  declared as `var x T` with no constructor.
- `Example` functions for anything shown in the README — they are compiled and
  run by `go test`, so the documentation cannot rot.

## API design rules

These are settled decisions, not preferences:

- **Concrete types, not interfaces.** Public types are structs. Interfaces
  cannot declare generic methods, which would forbid `Map[U]`.
- **The zero value must be valid and meaningful** — `Option[T]{}` is `None`,
  never a panic.
- **The library never panics.** No `panic()`, no `os.Exit`, no logging.
  Errors and absence are values.
- **Every container is a functor.** `Map` must be able to change the type
  parameter: `func (o Option[T]) Map[U any](f func(T) U) Option[U]`.
- **Every container is readable.** No type ships without a way to get its
  contents back out (`Get`, `Fold`/`Match`, `All() iter.Seq[T]`).
- **Immutability.** Operations return new values; nothing is mutated in place.
- **Integrate with Go, not against it.** Provide conversions to and from
  `(T, bool)`, `(T, error)`, `*T`, `[]T`, `iter.Seq[T]`, plus `json.Marshaler`
  and `sql.Scanner` where they make sense.
- **Complexity is part of the contract.** Document the cost of every operation
  in its godoc. No accidental O(n²) — add a benchmark when it is not obvious.
- **Ask before adding a new exported symbol.** The public surface is the
  expensive part; propose it before implementing it.

## Code style

- `gofmt` is mandatory — `gofmt -l .` must print nothing.
- **After modifying any Go code, run `golangci-lint run ./...` and fix every
  reported issue before considering the task complete.** Config is in
  `.golangci.yml` (golangci-lint v2). `go vet ./...` must also be clean.
- `depguard` rejects any non-standard-library import under `api/`. If it fires,
  the answer is to remove the import, not to add an exception.
- **Every exported symbol has a godoc comment**, starting with its name.
- Import order: stdlib, third-party, local module.
- Comments explain *why*, not *what*. No commented-out code, no `TODO` left
  behind — open an issue or do it.
- No dead code, no unused struct fields. If a field is scaffolding for a future
  feature, it does not get committed.
- Prefer clarity over cleverness: this is a library people read to learn.

## Documentation

- The README is a contract. If a change makes a README snippet wrong, fix the
  README **in the same commit**.
- README snippets must be backed by an `Example` test that compiles and runs.
- Any deviation from Vavr's semantics is documented in the godoc with a one-line
  rationale.

## Git

- **Never commit, push, tag, or force-push without an explicit instruction.**
- Never `git push --force` on `main`.
- Work on a branch; `main` is not a working branch.
- One logical change per commit. Imperative mood, present tense
  (`fix Try losing its error cause`).
- Do not amend or rebase commits you did not create in the current session.

## Issue and PR guidelines

- Never create an issue or pull request without the user's explicit instruction.
- If asked to open a PR on the user's behalf, state in the PR body that the
  change was prepared by an AI agent and may not have been independently
  reviewed or tested.

## Reporting

- Report outcomes honestly. If tests fail, show the output. If a step was
  skipped or a part of the task was left undone, say which and why.
- Never claim a change works without having run the tests.
- If a task is blocked by a decision that is not yours to make (a dependency, a
  breaking API change, a Go version bump), stop and ask rather than guessing.
