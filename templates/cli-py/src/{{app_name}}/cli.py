"""CLI entry point."""
import typer

app = typer.Typer(
    name="{{app_name}}",
    help="{{description}}",
    add_completion=False,
)


@app.callback(invoke_without_command=True)
def main(
    version: bool = typer.Option(
        False, "--version", "-V",
        help="Show version",
    ),
    fmt: str = typer.Option(
        "text", "--format", "-f",
        help="Output format (text, json, yaml)",
    ),
    verbose: bool = typer.Option(
        False, "--verbose", "-v",
        help="Verbose output",
    ),
) -> None:
    """{{description}}"""
    if version:
        from . import __version__
        typer.echo(f"{{app_name}} {__version__}")
        raise typer.Exit()


@app.command()
def hello(name: str = "World") -> None:
    """Say hello."""
    typer.echo(f"Hello, {name}!")
