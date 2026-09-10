"""Experimental cli helpers."""

import argparse
import logging
import sys
from . import read_blocks


def parse_args(args):
    parser = argparse.ArgumentParser(prog="rmscene")
    parser.add_argument("file", type=argparse.FileType("rb"), help="filename to read")
    return parser.parse_args(args)


def pprint_file(args) -> None:
    import pprint

    result = read_blocks(args.file)
    for el in result:
        print()
        pprint.pprint(el)


if __name__ == "__main__":
    args = parse_args(sys.argv[1:])
    logging.basicConfig(level=logging.INFO)
    pprint_file(args)
