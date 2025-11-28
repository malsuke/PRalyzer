# Project Development Guidelines

This document defines the core development philosophy, process, and Go-oriented coding standards for this project. The development agent is expected to strictly adhere to these guidelines at all times. The philosophy is heavily inspired by the teachings of Takuya Wada (t-wada) on Test-Driven Development (TDD).

## 1. Core Philosophy: Test-Driven
- **Tests Drive Development:** All production code is written only to make a failing test pass. Tests are not an afterthought; they are the specification and the driver of design.
- **Confidence in Refactoring:** A comprehensive test suite is our safety net. It allows us to refactor and improve the codebase fearlessly and continuously.
- **Testability Equals Good Design:** If Go code is difficult to test, it is a sign of poor design. Prioritize pure functions, dependency injection, and small interfaces so that components remain loosely coupled and highly cohesive.

## 2. The Development Cycle: Red-Green-Refactor-Commit
Follow this iterative cycle for every change, no matter how small. Explicitly state which phase you are in when sharing intermediate results.

### Phase 1: Red - Write a Failing Test
- **Goal:** Clearly define what needs to be accomplished.
- **Action:** Before writing implementation code, add a focused Go test (using the standard `testing` package or table-driven tests) that verifies a single piece of desired functionality.
- **Condition:** The new test must fail (**RED**) because the implementation does not exist yet. Run `go test ./...` to confirm the failure.

### Phase 2: Green - Make the Test Pass
- **Goal:** Fulfill the requirements defined by the failing test.
- **Action:** Write the absolute minimum amount of Go code needed to make the test pass (**GREEN**). Prefer small, composable functions with clear responsibilities.
- **Condition:** Do not add extra functionality. Confirm the test suite passes with `go test ./...`.

### Phase 3: Refactor - Improve the Design
- **Goal:** Clean up the code while keeping all tests green.
- **Action:** With the safety net of passing tests, improve the internal structure of the code. This includes, but is not limited to:
  - Removing duplication (DRY principle).
  - Improving names and exported symbol comments for clarity.
  - Simplifying complex logic, leveraging idiomatic Go patterns (e.g., guard clauses, small interfaces).
  - Ensuring all coding standards listed below are met.
- **Condition:** All tests must remain **GREEN** (verified via `go test ./...`) throughout the refactoring process.

### Phase 4: Commit - Save the Progress
- **Goal:** Record a functioning, small unit of work as a secure checkpoint.
- **Action:** After refactoring is complete and a final check confirms all tests are green, execute `git add .` to stage the changes. This serves as a stable checkpoint before proceeding to the next development cycle.
- **Condition:** The changes implemented in this cycle should represent a single, meaningful unit of work. The commit message should concisely describe this work.

## 3. Strict Coding Standards & Prohibitions

### 【CRITICAL】 No Hard-coding
Any form of hard-coded value is strictly forbidden.

- **Magic Numbers:** Do not use numeric literals directly in logic. Define them as named constants with `const`.
  - *Bad:* `if age > 20`
  - *Good:* `const adultAge = 20; if age > adultAge`
- **Configuration Values:** API keys, URLs, file paths, and other environmental settings must be sourced from configuration layers (e.g., environment variables via `os.Getenv`, configuration structs, or the existing `internal/config` package). They must never be committed directly in Go source files.
- **User-facing Strings:** Text for UI, logs, or errors should be centralized via constants or localization helpers to ease maintenance and internationalization.

### Other Key Standards (Go Focused)
- **Single Responsibility Principle (SRP):** Every package, type, or function should handle a single responsibility. Split packages when concerns diverge.
- **DRY (Don't Repeat Yourself):** Avoid code duplication. Extract shared logic into helper functions or internal packages.
- **Clear and Intentional Naming:** Follow Go naming conventions.
  - 関数・メソッド名は動詞（あるいは動詞句）で開始し、動作が伝わるようにする（例：`LoadConfig`、`handleRequest`）。
  - 変数は用途がわかる短い lowerCamelCase 名を使う。パッケージ内でのみ意味が通じる略語は避ける。
  - 定数名は `PascalCase`（エクスポート）または `camelCase`（非エクスポート）とし、意味を明確にする。
  - 型名は名詞で、責務が一目でわかるものにする。インターフェース名には機能を表す名詞＋`er`/`Provider` などを用いる。
  - Exported identifiers require doc comments in full sentences. Prefer short, descriptive names.
- **Guard Clauses / Early Return:** Prefer early returns to avoid deeply nested `if` ladders.
- **Error Handling:** Always check errors. Wrap with `%w` when adding context. Return sentinel errors or use `errors.Join`/`errors.Is` and `errors.As` as appropriate.
- **Context Propagation:** Functions performing I/O must accept a `context.Context`. Do not store contexts in structs.
- **Concurrency Safety:** Use goroutines and channels judiciously. Protect shared state with appropriate synchronization primitives.
- **Security First:** Treat all external input as untrusted. Validate and sanitize inputs, and follow secure defaults for HTTP clients, file handling, and cryptographic operations.

## 4. Go Tooling & Quality Checks
- **Formatting:** Run `gofmt -w` (or `goimports` if available) on all modified files. Formatting is non-negotiable.
- **Static Analysis:** Run `go vet ./...` to catch suspicious constructs. Use `staticcheck ./...` when available to enforce broader lint rules.
- **Linting:** If `golangci-lint` is configured, execute `golangci-lint run ./...` before submitting changes.
- **Testing:** Execute `go test ./...` frequently. For targeted runs use `go test ./path/to/pkg -run TestName`.
- **Race Detection:** When concurrency is in play, run `go test -race ./...`.
- **Module Hygiene:** Keep `go.mod` and `go.sum` tidy via `go mod tidy` when dependencies change. Never commit extraneous modules or replace directives without justification.

## 5. Additional Go Best Practices
- **Package Design:** Prefer small, focused packages. Avoid import cycles. Internal packages protect implementation details.
- **Interfaces:** Define interfaces on the consumer side. Keep interfaces small and focused (one or two methods).
- **Struct Initialization:** Use constructor functions when invariants must be enforced. Validate inputs and return descriptive errors.
- **Dependency Injection:** Accept dependencies as interfaces or function parameters to facilitate testing.
- **Logging:** Use consistent logging primitives. Avoid excessive logging noise. Ensure logs are structured when possible.
- **Documentation:** Maintain README and package-level documentation (`doc.go`) so newcomers can understand intent quickly.

These guidelines ensure that Go code in this project remains maintainable, testable, and aligned with best-in-class development practices.***