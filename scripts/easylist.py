#!/usr/bin/env python3
import sys
import os


def main():
    filename = sys.argv[1]
    result = []
    with open(filename) as fobj:
        for line in fobj:
            if line.startswith("||"):
                line = line.strip("|\n ^")
                words = line.split("^")
                url = words[0]
                domains = url.split("/")
                domain = domains[0]
                result.append(domain)


    # now write back
    newname = os.path.basename(filename)
    with open(f"blocklists/clean-{newname}", "w") as fobj:
        for line in result:
            fobj.write(f"{line}\n")
    print(f"Wrote: blocklists/clean-{newname}")

if __name__ == "__main__":
    main()
