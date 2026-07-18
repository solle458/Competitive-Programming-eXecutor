"""Monkey-patch for online-judge-api-client to support AtCoder's MiB memory limit format.

AtCoder changed the problem page format from "Memory Limit: 1024 MB" to
"Memory Limit: 1024 MiB" (Mebibytes). online-judge-api-client's regex only
matches KB|MB, causing AssertionError when parsing newer problem pages and
the contest tasks list table.

This module patches onlinejudge.service.atcoder.AtCoderProblemData._from_html
and _from_table_row to support MiB before any oj command runs.

Ported from ac-jj (https://github.com/solle458/ac-jj).
"""
import re
from typing import Any

import bs4


def _parse_memory_limit_from_labelled_text(memory_limit: str) -> int:
    """Parse labelled memory limit text (problem page format) to bytes."""
    parsed = re.search(
        r"^(メモリ制限|Memory Limit): ([0-9.]+) (KB|MB|KiB|MiB)", memory_limit
    )
    assert parsed, f"Memory limit regex did not match: {memory_limit!r}"
    return _memory_limit_value_to_bytes(parsed.group(2), parsed.group(3))


def _parse_memory_limit_from_table_cell(memory_limit: str) -> int:
    """Parse memory limit table cell text (tasks list format) to bytes."""
    for suffix, unit in (
        (" KB", "KB"),
        (" MB", "MB"),
        (" KiB", "KiB"),
        (" MiB", "MiB"),
    ):
        if memory_limit.endswith(suffix):
            return _memory_limit_value_to_bytes(
                memory_limit.removesuffix(suffix), unit
            )
    raise AssertionError(f"Unsupported memory limit table cell: {memory_limit!r}")


def _memory_limit_value_to_bytes(value: str, unit: str) -> int:
    """Convert a memory limit value and unit to bytes."""
    if unit == "KB":
        return int(float(value) * 1000)
    if unit == "MB":
        return int(float(value) * 1000 * 1000)
    if unit == "KiB":
        return int(float(value) * 1024)
    if unit == "MiB":
        return int(float(value) * 1024 * 1024)
    raise AssertionError(f"Unsupported memory limit unit: {unit!r}")


def _parse_time_limit_msec(time_limit: str) -> int:
    """Parse time limit text to milliseconds."""
    from onlinejudge._implementation import utils

    for time_limit_prefix in ("実行時間制限: ", "Time Limit: "):
        if time_limit.startswith(time_limit_prefix):
            body = utils.remove_prefix(time_limit, time_limit_prefix)
            if body.endswith(" msec"):
                return int(utils.remove_suffix(body, " msec"))
            if body.endswith(" sec"):
                return int(float(utils.remove_suffix(body, " sec")) * 1000)
            break
    raise AssertionError(f"Unsupported time limit format: {time_limit!r}")


def _parse_time_limit_msec_from_table_cell(time_limit: str) -> int:
    """Parse time limit table cell text to milliseconds."""
    from onlinejudge._implementation import utils

    if time_limit.endswith(" msec"):
        return int(utils.remove_suffix(time_limit, " msec"))
    if time_limit.endswith(" sec"):
        return int(float(utils.remove_suffix(time_limit, " sec")) * 1000)
    raise AssertionError(f"Unsupported time limit table cell: {time_limit!r}")


def _patched_from_html(
    cls: Any,
    html: bytes,
    *,
    problem: Any,
    session: Any = None,
    response: Any = None,
    timestamp: Any = None,
) -> Any:
    """Patched _from_html supporting KB, MB, KiB, and MiB for memory limit."""
    from onlinejudge._implementation import utils
    from onlinejudge.service.atcoder import AtCoderProblemData

    soup = bs4.BeautifulSoup(html, utils.HTML_PARSER)
    h2 = soup.find("span", class_="h2")
    if h2 is None:
        raise AssertionError("Problem page: span.h2 not found (HTML structure may have changed)")

    alphabet, _, name = utils.get_direct_children_text(h2).strip().partition(" - ")

    limit_p = h2.find_next_sibling("p")
    if limit_p is None:
        raise AssertionError(
            "Problem page: limit <p> not found after h2 (HTML structure may have changed)"
        )
    time_limit, memory_limit = limit_p.text.strip().split(" / ")
    time_limit_msec = _parse_time_limit_msec(time_limit)
    memory_limit_byte = _parse_memory_limit_from_labelled_text(memory_limit)

    return AtCoderProblemData(
        alphabet=alphabet,
        html=html,
        memory_limit_byte=memory_limit_byte,
        name=name,
        problem=problem,
        response=response,
        session=session,
        time_limit_msec=time_limit_msec,
        timestamp=timestamp,
    )


def _patched_from_table_row(
    cls: Any,
    tr: bs4.Tag,
    *,
    session: Any,
    response: Any,
    timestamp: Any,
) -> Any:
    """Patched _from_table_row supporting KB, MB, KiB, and MiB for memory limit."""
    from onlinejudge.service.atcoder import AtCoderProblem, AtCoderProblemData

    tds = tr.find_all("td")
    assert 4 <= len(tds) <= 5
    path = tds[1].find("a")["href"]
    problem = AtCoderProblem.from_url("https://atcoder.jp" + path)
    assert problem is not None
    alphabet = tds[0].text
    name = tds[1].text
    time_limit_msec = _parse_time_limit_msec_from_table_cell(tds[2].text)
    memory_limit_byte = _parse_memory_limit_from_table_cell(tds[3].text)
    if len(tds) == 5:
        assert tds[4].text.strip() in ("", "Submit", "提出")

    return AtCoderProblemData(
        alphabet=alphabet,
        memory_limit_byte=memory_limit_byte,
        name=name,
        problem=problem,
        response=response,
        session=session,
        time_limit_msec=time_limit_msec,
        timestamp=timestamp,
    )


def apply() -> None:
    """Apply MiB support patches to onlinejudge.service.atcoder."""
    import onlinejudge.service.atcoder as atcoder_module

    atcoder_module.AtCoderProblemData._from_html = classmethod(_patched_from_html)
    atcoder_module.AtCoderProblemData._from_table_row = classmethod(_patched_from_table_row)
