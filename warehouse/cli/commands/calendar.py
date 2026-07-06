"""Calendar CLI — ICS import, query, and data-quality checks."""

from __future__ import annotations

from pathlib import Path

import typer
from rich.console import Console
from rich.table import Table

from warehouse.calendar.import_pipeline import (
    get_calendar_stats,
    import_ics_files,
    query_occurrences,
    run_calendar_doctor,
)
from warehouse.core import AppSettings
from warehouse.core.errors import ConfigError

app = typer.Typer(help="Calendar import and query tools.")
console = Console()


def _settings() -> AppSettings:
    try:
        s = AppSettings.from_config()
        s.ensure_database()
    except ConfigError as exc:
        console.print(f"[red]{exc}[/red]")
        raise typer.Exit(1) from exc
    return s


@app.command("import")
def import_calendar(
    path: str = typer.Argument(..., help="Directory containing .ics files to import."),
    source_system: str | None = typer.Option(
        None, "--source", help="Override source system (google, apple, outlook, todoist)."
    ),
    future_years: int = typer.Option(
        2, "--future-years", help="Years forward to expand occurrences."
    ),
    past_years: int = typer.Option(1, "--past-years", help="Years back to expand occurrences."),
) -> None:
    """Import ICS calendar files and expand occurrences."""
    settings = _settings()
    import_path = Path(path)
    if not import_path.exists():
        console.print(f"[red]Path not found: {path}[/red]")
        raise typer.Exit(1)

    console.print(f"Scanning {import_path} for .ics files...")

    result = import_ics_files(
        settings.database_path,
        import_path,
        source_system=source_system,
        future_years=future_years,
        past_years=past_years,
    )

    table = Table(title="Calendar Import")
    table.add_column("Metric", style="cyan")
    table.add_column("Value", style="green", justify="right")
    table.add_row("Batch ID", result["batch_id"])
    table.add_row("Files found", str(result["file_count"]))
    table.add_row("Events imported", str(result["event_count"]))
    table.add_row("Warnings", str(result["warning_count"]))
    table.add_row("Errors", str(result["error_count"]))
    table.add_row("Occurrences expanded", str(result["occurrence_count"]))
    table.add_row("Expansion window", result["expand_window"])
    console.print(table)

    if result["errors"]:
        console.print("\n[red]Errors:[/red]")
        for path_str, msg in result["errors"][:10]:
            console.print(f"  {path_str}: {msg}")


@app.command("query")
def query_calendar(
    text: str = typer.Argument(
        ..., help="Search term for event summaries, descriptions, and locations."
    ),
    from_date: str | None = typer.Option(None, "--from", help="Filter from ISO date (YYYY-MM-DD)."),
    to_date: str | None = typer.Option(None, "--to", help="Filter to ISO date (YYYY-MM-DD)."),
    limit: int = typer.Option(50, "--limit", "-n", help="Max results."),
) -> None:
    """Search calendar occurrences by text."""
    settings = _settings()
    results = query_occurrences(
        settings.database_path,
        text,
        from_date=from_date,
        to_date=to_date,
        limit=limit,
    )

    if not results:
        console.print("[yellow]No matching events found.[/yellow]")
        return

    table = Table(title=f'Calendar Search: "{text}"')
    table.add_column("Date", style="cyan")
    table.add_column("Summary", style="white")
    table.add_column("Source", style="dim")
    for r in results:
        date_display = r["date"] or (r["start_utc"][:10] if r["start_utc"] else "—")
        cancelled = " [CANCELLED]" if r["is_cancelled"] else ""
        table.add_row(
            date_display,
            (r["summary"] or "(no summary)") + cancelled,
            f"{r['source_system']}/{r['source_file']}",
        )
    console.print(table)
    console.print(f"\n[dim]{len(results)} results (limit: {limit})[/dim]")


@app.command("stats")
def calendar_stats() -> None:
    """Show calendar import statistics."""
    settings = _settings()
    stats = get_calendar_stats(settings.database_path)

    table = Table(title="Calendar Statistics")
    table.add_column("Metric", style="cyan")
    table.add_column("Value", style="green", justify="right")
    table.add_row("Raw events", str(stats["raw_events"]))
    table.add_row("Occurrences", str(stats["occurrences"]))
    table.add_row("Import batches", str(stats["batches"]))
    table.add_row("Recurring events", str(stats["recurring"]))
    table.add_row("Cancelled occurrences", str(stats["cancelled"]))
    console.print(table)

    if stats["by_source_system"]:
        console.print("\n[bold]By Source System:[/bold]")
        for system, count in stats["by_source_system"].items():
            console.print(f"  {system}: {count}")


@app.command("doctor")
def calendar_doctor() -> None:
    """Run data-quality checks on calendar data."""
    settings = _settings()
    findings = run_calendar_doctor(settings.database_path)

    if not findings:
        console.print("[green]No issues found.[/green]")
        return

    for f in findings:
        label = f["severity"].upper()
        color = {"error": "red", "warn": "yellow", "info": "dim"}.get(f["severity"], "")
        count_str = f" count={f['count']:,}" if "count" in f else ""
        console.print(f"[{color}]{label:5s}[/{color}] {f['check']:28s}{count_str}  {f['detail']}")
