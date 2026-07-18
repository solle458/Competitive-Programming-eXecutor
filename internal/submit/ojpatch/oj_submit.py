"""Entry point for MiB-patched oj submit.

Applies the memory-limit patch before running oj submit.
Invoked by cpx as: <oj-python> oj_submit.py --yes <url> <file>
"""
import sys
from pathlib import Path

# Ensure sibling oj_patch.py is importable when run as a script.
sys.path.insert(0, str(Path(__file__).resolve().parent))

from oj_patch import apply

apply()

sys.argv = ["oj", "submit"] + sys.argv[1:]
from onlinejudge_command.main import main

main()
