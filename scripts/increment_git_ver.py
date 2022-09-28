#!/usr/bin/python3
import logging
import sys
import subprocess
from packaging import version
import argparse


parser = argparse.ArgumentParser()
parser.add_argument("--debug", help="Add debug printing",
                    default=False, required=False)
args = parser.parse_args()

if args.debug:
    logging.basicConfig(stream=sys.stdout, level=logging.DEBUG, )

logging.debug("DEBUG")


def current_ver() -> str():
    # Get the two most recent versions
    out = subprocess.check_output(
        ["git", "tag", "--sort=-version:refname"]).decode('ascii').strip()
    out_arr = out.split('\n')

    logging.debug(out_arr[0])
    logging.debug(out_arr[1])
    logging.debug(out_arr[2])

    return out_arr[0]


def current_branch() -> str:
    return subprocess.check_output(["git", "rev-parse", "--abbrev-ref", "HEAD"]).decode('ascii').strip()


subprocess.check_output(["git", "pull", "--quiet", "--tags"])
current_version = version.parse(current_ver())
logging.debug("selected version: %s", current_version)

current_branch = current_branch()
next_version = ""

# fixedString is needed for some repos with borked tags, i.e. 0.0.1.2.3
fixedString = "{major}.{minor}.{micro}".format(
    major=current_version.major,
    minor=current_version.minor,
    micro=current_version.micro
)

# If this is not a pre release then increment the microversion
if not current_version.is_prerelease:
    fixedString = "{major}.{minor}.{micro}".format(
        major=current_version.major,
        minor=current_version.minor,
        micro=current_version.micro + 1
    )


if current_branch == 'develop':
    if current_version.is_prerelease:
        next_version = "v{baseVer}-rc.{rc:02d}".format(
            baseVer=fixedString,
            rc=current_version.pre[1] + 1
        )
    else:
        next_version = "v{baseVer}-rc.01".format(baseVer=fixedString)

else:
    print(
        "invalid branch, should be develop: current is " + current_branch)
    exit(1)

logging.debug("%s -> %s", current_version, next_version)
print(next_version)
