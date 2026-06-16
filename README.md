# Be - a minimal test assertions package

If you want simple test assertions and feel like [testify](https://pkg.go.dev/github.com/attic-labs/testify/assert) is too much, but [is](https://pkg.go.dev/github.com/matryer/is) is too basic, you might like `be`.

Highlights:

-   Minimal API: `Equal`, `Err`, and `True` assertions.
-   Correctly compares `time.Time` values and other types with an `Equal` method.
-   Flexible error assertions: check if an error exists, check its value, type, or any combination of these.
-   Zero hassle.

Be is new, but it's ready for production (or maybe I should say "testing" :) I've used it in three very different projects — a CLI tool, an API server, and a database engine — and it worked great every time.

## Usage

Install with go get:

```text
go get github.com/nalgeon/be
```

`Equal` asserts that two values are equal:

```go
func Test(t *testing.T) {
    t.Run("pass", func(t *testing.T) {
        got, want := "hello", "hello"
        be.Equal(t, got, want)
        // ok
    })

    t.Run("fail", func(t *testing.T) {
        got, want := "olleh", "hello"
        be.Equal(t, got, want)
        // got: "olleh"; want: "hello"
    })
}
```

Or that a value matches any of the given values:

```go
func Test(t *testing.T) {
    got := 2 * 3 * 7
    be.Equal(t, got, 21, 42, 84)
    // ok
}
```

`Err` asserts that there is an error:

```go
func Test(t *testing.T) {
    _, err := regexp.Compile("he(?o") // invalid
    be.Err(t, err)
    // ok
}
```

Or that there are no errors:

```go
func Test(t *testing.T) {
    _, err := regexp.Compile("he??o") // valid
    be.Err(t, err, nil)
    // ok
}
```

Or that an error message contains a substring:

```go
func Test(t *testing.T) {
    _, err := regexp.Compile("he(?o") // invalid
    be.Err(t, err, "invalid or unsupported")
    // ok
}
```

Or that an error matches the expected error according to `errors.Is`:

```go
func Test(t *testing.T) {
    err := &fs.PathError{
        Op: "open",
        Path: "file.txt",
        Err: fs.ErrNotExist,
    }
    be.Err(t, err, fs.ErrNotExist)
    // ok
}
```

Or that the error type matches the expected type according to `errors.As`:

```go
func Test(t *testing.T) {
    got := &fs.PathError{
        Op: "open",
        Path: "file.txt",
        Err: fs.ErrNotExist,
    }
    be.Err(t, got, reflect.TypeFor[*fs.PathError]())
    // ok
}
```

Or a mix of the above:

```go
func Test(t *testing.T) {
    err := AppError("oops")
    be.Err(t, err,
        "failed",
        AppError("oops"),
        reflect.TypeFor[AppError](),
    )
    // ok
}
```

`True` asserts that an expression is true:

```go
func Test(t *testing.T) {
    s := "go is awesome"
    be.True(t, len(s) > 0)
    // ok
}
```

That's it!

## Design decisions

Be is opinionated. It only has three assert functions, which are perfectly enough to write good tests.

Unlike other testing packages, Be doesn't support custom error messages. When a test fails, you'll end up checking the code anyway, so why bother? The line number shows the way.

Be has flexible error assertions. You don't need to choose between `Error`, `ErrorIs`, `ErrorAs`, `ErrorContains`, `NoError`, or anything like that — just use `be.Err`. It covers everything.

Be doesn't fail the test when an assertion fails, so you can see all the errors at once instead of hunting them one by one. The only exception is when the `be.Err(err, nil)` assertion fails — this means there was an unexpected error. In this case, the test terminates immediately because any following assertions probably won't make sense and could cause panics.

The parameter order is (got, want), not (want, got). It just feels more natural — like saying "account balance is 100 coins" instead of "100 coins is the account balance".

Be has ≈150 lines of code (+500 lines for tests). For comparison, `is` has ≈250 loc (+250 lines for tests).

## Contributing

Bug fixes are welcome. For anything other than bug fixes, please open an issue first to discuss your proposed changes. The package has a very limited scope, so it's important to discuss any new features before implementing them.

Make sure to add or update tests as needed.

## License

Created by [Anton Zhiyanov](https://antonz.org/). Released under the MIT License.
