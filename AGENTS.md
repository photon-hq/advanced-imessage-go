# AGENTS.md

Guidance for agents working on `advanced-imessage-go`. This is a Go **library**, so the public API is the product — getting it right matters more than internal cleverness, because backwards-compatibility makes mistakes expensive after 1.0.

## Design principles

Based on Abhinav Gupta's [Designing Go Libraries](https://abhinavg.net/2022/12/06/designing-go-libraries/). There are exceptions to everything here — apply judgment, but justify deviations.

### Maximize four axes
- **Usability** — make common tasks easy, uncommon tasks possible. Easy to use correctly, hard to use incorrectly.
- **Readability** — code is read far more than written. Clear, descriptive names; clean APIs.
- **Flexibility** — well-defined abstractions that allow extension and third-party implementations.
- **Testability** — let users swap components for fakes. Plan for it at design time, not after.

### Backwards compatibility
- **Interfaces are forever.** Adding a method to an exported interface breaks every external implementer; removing/changing signatures breaks callers. Once shipped in a stable release, an interface cannot change.
- **Return concrete structs, accept interfaces.** Structs let you add methods compatibly; users can still declare their own interfaces for mocking.
- **SemVer honestly.** Pre-1.0 is the only cheap window for breaking changes. After 1.0, breaking changes require a MAJOR bump.
- **Support current + previous Go minor** only.
- Breaking changes (per Go 1 compat): renaming/removing exported entities, changing types, modifying signatures, adding methods to exported interfaces, behavioral changes that violate contracts.

### API surface discipline
- **Work backwards** — write the consuming pseudo-code and docs before implementing. Exercise all four axes against the sketch.
- **Minimize surface area** — "when in doubt, leave it out." Smaller surface = more freedom to refactor. Never export provisional "maybe" APIs.
- **Use `internal/`** liberally for helpers and non-core functionality. Never expose internal types through an exported API.
- **Avoid unknown outputs** (Hyrum's Law — observable behavior becomes a de facto contract):
  - Validate inputs strictly; loosen later if needed.
  - Clone slices/maps at API boundaries to guard internal state.
  - Return zero values on error (except documented partial-result cases like `Writer.Write`).
  - Specify ordering in contracts, or sort at the boundary.
- **No global/package-level state.** Encapsulate state in types; convert top-level functions to methods.

### Abstraction patterns
- **Accept, don't instantiate.** Take `io.Reader` over a filename. Offer convenience constructors (`NewFromFile`) for common cases.
- **Accept interfaces** for complex dependencies; declare custom interfaces even for types you don't own, with a `var _ Iface = (*concrete)(nil)` compile-time check.
- **Upcasting** for optional behaviors without changing an interface (the `io.WriteString` → `io.StringWriter` pattern): define a new interface embedding the old + new methods, type-assert, fall back gracefully.
- **Parameter objects / result objects** for functions with 3+ params or returns — adding optional fields stays compatible.
- **Functional options** (opaque `Option` with an unexported `apply` method) for sparse optional config when there are ≤2 required params. Never mix functional options and a parameter object in the same function.

### Behavior & semantics
- **Goroutines:** don't grow unbounded (never one-per-slice-item; use bounded worker pools), don't leak (always provide shutdown; test with `goleak`). "Don't be greedy and clean up after yourself."
- **Errors:** use modern `errors.Is` / `errors.As` / `%w` wrapping. Don't log-and-return. Don't export error types unnecessarily.
- **Reflection:** use `reflect` sparingly and only in performance-insensitive paths; it panics on bad input and tests don't guarantee correctness.

### Naming & documentation
- **No `FooManager`** — find the real noun (`RoundTripper`, `ServeMux`, `FlagSet`, `sql.DB`).
- **No kitchen-sink packages** (`common`, `util`). Organize by responsibility.
- **Qualified package names** — `httptest` not `test`, `zapcore` not `core`, to avoid forcing named imports.
- **Document for users, not maintainers** — explain purpose, usage, and relationships; avoid implementation details and non-documenting docs ("RequestHandler handles requests"). Prefer examples (example tests for tested samples).
- **Keep a changelog** separate from commit logs (keepachangelog.com format); track user-facing changes as they land.
