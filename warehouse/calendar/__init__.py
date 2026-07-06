"""ICS file parser — ported from toolbox/apps/filekit/src/cal.rs."""

from __future__ import annotations

import hashlib
import json
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path

from dateutil import parser as dateutil_parser
from dateutil import tz


@dataclass
class RawEvent:
    """A parsed ICS VEVENT before database insertion."""

    import_batch_id: str
    source_system: str | None = None
    source_file: str = ""
    source_path: str | None = None
    source_hash: str | None = None
    uid: str = ""
    recurrence_id_raw: str | None = None
    recurrence_id_utc: str | None = None
    sequence: int | None = None
    dtstamp_utc: str | None = None
    created_utc: str | None = None
    last_modified_utc: str | None = None
    calendar_name: str | None = None
    prodid: str | None = None
    method: str | None = None
    event_type: str | None = None
    classification: str | None = None
    status: str | None = None
    transp: str | None = None
    summary: str | None = None
    description: str | None = None
    location: str | None = None
    dtstart_raw: str | None = None
    dtstart_tzid: str | None = None
    dtstart_utc: str | None = None
    dtstart_kind: str = "unknown"
    dtend_raw: str | None = None
    dtend_tzid: str | None = None
    dtend_utc: str | None = None
    dtend_kind: str = "unknown"
    duration_raw: str | None = None
    all_day: bool = False
    rrule_raw: str | None = None
    exdate_raw: str | None = None
    rdate_raw: str | None = None
    exrule_raw: str | None = None
    tzid: str | None = None
    organizer: str | None = None
    attendees_json: str | None = None
    categories_json: str | None = None
    url: str | None = None
    raw: str = ""
    parse_warnings_json: str | None = None
    parse_error: str | None = None
    ingested_at: str = ""


@dataclass
class Occurrence:
    """An expanded event occurrence."""

    raw_event_id: int
    uid: str
    occurrence_key: str
    recurrence_id_utc: str | None = None
    occurrence_start_utc: str = ""
    occurrence_end_utc: str | None = None
    occurrence_date: str | None = None
    is_all_day: bool = False
    is_generated: bool = False
    is_override: bool = False
    is_cancelled: bool = False
    is_excluded: bool = False
    status: str | None = None
    summary: str | None = None
    description: str | None = None
    location: str | None = None
    expansion_window_start: str | None = None
    expansion_window_end: str | None = None
    expanded_at: str = ""
    expansion_version: str = ""


def _file_hash(path: Path) -> str:
    """SHA-256 hash of a file for deduplication."""
    hasher = hashlib.sha256()
    with open(path, "rb") as f:
        while True:
            chunk = f.read(8192)
            if not chunk:
                break
            hasher.update(chunk)
    return hasher.hexdigest()


def _unfold_ics(content: str) -> str:
    """Unfold continuation lines in ICS content."""
    result: list[str] = []
    for line in content.splitlines():
        line = line.rstrip("\r")
        if not line:
            continue
        if result and (line.startswith(" ") or line.startswith("\t")):
            result[-1] += line.lstrip(" \t")
        else:
            result.append(line)
    return "\n".join(result)


def _parse_ical_value(raw: str) -> str:
    """Decode escaped ICS text values."""
    return raw.replace("\\n", "\n").replace("\\,", ",").replace("\\;", ";").replace("\\\\", "\\")


def _parse_datetime(
    raw: str, default_tz_name: str | None = None
) -> tuple[str | None, str | None, str]:
    """Parse an ICS datetime value. Returns (raw, utc, kind)."""
    raw = raw.strip()

    # Date-only: VALUE=DATE:YYYYMMDD
    if "VALUE=DATE" in raw:
        if ":" in raw:
            date_str = raw.split(":", 1)[1].strip()
        else:
            date_str = raw
        if len(date_str) == 8 and date_str.isdigit():
            return date_str, None, "date"
        return raw, None, "date"

    # Plain date: YYYYMMDD
    if len(raw) == 8 and raw.isdigit():
        return raw, None, "date"

    # UTC: ends with Z
    if raw.endswith("Z"):
        try:
            dt = dateutil_parser.isoparse(raw)
            if dt.tzinfo is None:
                dt = dt.replace(tzinfo=timezone.utc)
            return raw, dt.astimezone(timezone.utc).isoformat(), "utc"
        except (ValueError, OverflowError):
            pass

    # Extract TZID parameter and value
    parts = raw.split(":", 1)
    params_part = parts[0] if parts else ""
    value_part = parts[1] if len(parts) > 1 else raw

    tzid_param = None
    for param in params_part.split(";"):
        if param.startswith("TZID="):
            tzid_param = param[5:]
            break

    # Try parsing as ISO datetime
    try:
        dt = dateutil_parser.isoparse(value_part)
        if dt.tzinfo is not None:
            return raw, dt.astimezone(timezone.utc).isoformat(), "utc"
        return raw, dt.replace(tzinfo=timezone.utc).isoformat(), "floating"
    except (ValueError, OverflowError):
        pass

    # Try naive formats with timezone
    tz_name = tzid_param or default_tz_name
    naive_formats = ["%Y%m%dT%H%M%S", "%Y%m%dT%H%M"]
    for fmt in naive_formats:
        try:
            ndt = datetime.strptime(value_part, fmt)
            if tz_name:
                try:
                    zone = tz.gettz(tz_name)
                    if zone:
                        ndt = ndt.replace(tzinfo=zone)
                        return raw, ndt.astimezone(timezone.utc).isoformat(), "zoned"
                except Exception:
                    pass
            return raw, ndt.replace(tzinfo=timezone.utc).isoformat(), "floating"
        except ValueError:
            continue

    return raw, None, "unknown"


def _infer_source_system(prodid: str | None, path: Path) -> str:
    """Infer the source system from PRODID or filename."""
    if prodid:
        p_lower = prodid.lower()
        if "google" in p_lower:
            return "google"
        if "apple" in p_lower or "apple inc" in p_lower:
            return "apple"
        if "microsoft" in p_lower or "outlook" in p_lower:
            return "outlook"
        if "mimecast" in p_lower:
            return "mimecast"
        if "todoist" in p_lower:
            return "todoist"

    name = path.stem.lower()
    if "todoist" in name:
        return "todoist"
    if "mimecast" in name:
        return "mimecast"
    if "apple" in name:
        return "apple"
    return "unknown"


def parse_ics_file(
    path: Path,
    import_batch_id: str,
    inferred_source: str,
) -> list[RawEvent]:
    """Parse an ICS file into RawEvent structs."""
    content = path.read_text(encoding="utf-8", errors="replace")
    unfolded = _unfold_ics(content)
    source_hash = _file_hash(path)
    source_file = path.name or "unknown.ics"
    source_path = str(path)

    events: list[RawEvent] = []
    prodid: str | None = None
    method: str | None = None
    calendar_name: str | None = None
    current_tzid: str | None = None
    vevent_raw: list[str] = []
    fields: dict[str, str] = {}
    in_vevent = False

    for line in unfolded.splitlines():
        if not line.strip():
            continue

        if line == "BEGIN:VEVENT":
            in_vevent = True
            vevent_raw = [line]
            fields = {}
            current_tzid = None
            continue

        if line == "END:VEVENT":
            in_vevent = False
            vevent_raw.append(line)

            uid = fields.get("UID", "")
            if not uid:
                uid = f"MISSING-UID-{int(datetime.now(timezone.utc).timestamp() * 1_000_000)}"

            dtstamp_utc = _parse_datetime(fields.get("DTSTAMP", ""), None)[1]
            created_utc = _parse_datetime(fields.get("CREATED", ""), None)[1]
            last_modified_utc = _parse_datetime(fields.get("LAST-MODIFIED", ""), None)[1]

            dtstart_raw = fields.get("DTSTART")
            dtend_raw = fields.get("DTEND")

            # Determine dtstart kind and UTC
            dtstart_tzid = current_tzid
            if dtstart_raw:
                for param in dtstart_raw.split(";")[:-1]:
                    if param.startswith("TZID="):
                        dtstart_tzid = param[5:]
                        break

            dtstart_kind = "unknown"
            if dtstart_raw:
                if "VALUE=DATE" in dtstart_raw:
                    dtstart_kind = "date"
                elif dtstart_raw.endswith("Z"):
                    dtstart_kind = "utc"
                else:
                    dtstart_kind = "zoned"

            dtstart_utc = _parse_datetime(dtstart_raw or "", current_tzid)[1]
            dtend_utc = _parse_datetime(dtend_raw or "", current_tzid)[1]
            all_day = dtstart_kind == "date"

            # Parse recurrence fields
            recurrence_id_raw = fields.get("RECURRENCE-ID")
            recurrence_id_utc = _parse_datetime(recurrence_id_raw or "", current_tzid)[1]

            # Categories and attendees as JSON arrays
            categories_json = None
            if "CATEGORIES" in fields:
                cats = [c.strip() for c in fields["CATEGORIES"].split(",")]
                categories_json = json.dumps(cats)

            attendees_json = None
            if "ATTENDEE" in fields:
                atts = [a.strip() for a in fields["ATTENDEE"].split(",")]
                attendees_json = json.dumps(atts)

            warnings: list[str] = []
            if not fields.get("UID"):
                warnings.append("Missing UID")

            events.append(
                RawEvent(
                    import_batch_id=import_batch_id,
                    source_system=inferred_source,
                    source_file=source_file,
                    source_path=source_path,
                    source_hash=source_hash,
                    uid=uid,
                    recurrence_id_raw=recurrence_id_raw,
                    recurrence_id_utc=recurrence_id_utc,
                    sequence=int(fields["SEQUENCE"]) if fields.get("SEQUENCE") else None,
                    dtstamp_utc=dtstamp_utc,
                    created_utc=created_utc,
                    last_modified_utc=last_modified_utc,
                    calendar_name=calendar_name,
                    prodid=prodid,
                    method=method,
                    event_type=None,
                    classification=fields.get("CLASS"),
                    status=fields.get("STATUS"),
                    transp=fields.get("TRANSP"),
                    summary=_parse_ical_value(fields["SUMMARY"]) if fields.get("SUMMARY") else None,
                    description=_parse_ical_value(fields["DESCRIPTION"])
                    if fields.get("DESCRIPTION")
                    else None,
                    location=fields.get("LOCATION"),
                    dtstart_raw=dtstart_raw,
                    dtstart_tzid=dtstart_tzid,
                    dtstart_utc=dtstart_utc,
                    dtstart_kind=dtstart_kind,
                    dtend_raw=dtend_raw,
                    dtend_tzid=None,
                    dtend_utc=dtend_utc,
                    dtend_kind="zoned",
                    duration_raw=fields.get("DURATION"),
                    all_day=all_day,
                    rrule_raw=fields.get("RRULE"),
                    exdate_raw=fields.get("EXDATE"),
                    rdate_raw=fields.get("RDATE"),
                    exrule_raw=fields.get("EXRULE"),
                    tzid=current_tzid,
                    organizer=fields.get("ORGANIZER"),
                    attendees_json=attendees_json,
                    categories_json=categories_json,
                    url=fields.get("URL"),
                    raw="\n".join(vevent_raw),
                    parse_warnings_json=json.dumps(warnings) if warnings else None,
                    parse_error=None,
                    ingested_at=datetime.now(timezone.utc).isoformat(),
                )
            )
            continue

        if in_vevent:
            vevent_raw.append(line)

        if ":" not in line:
            continue

        key_full, _, value = line.partition(":")
        value = value.strip()
        key_base = key_full.split(";")[0]

        if key_base in ("BEGIN", "END"):
            continue

        if key_base == "TZID":
            current_tzid = value
            continue

        if in_vevent:
            if key_base in fields:
                fields[key_base] += "," + value
            else:
                fields[key_base] = value
        else:
            if key_base == "PRODID":
                prodid = value
            elif key_base == "METHOD":
                method = value
            elif key_base == "X-WR-CALNAME":
                calendar_name = value

    return events
