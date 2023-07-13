import argparse
import sys

import autograder.style

def run(args):
    count, total_lines = autograder.style.check_paths(args.paths)
    print("Found %d style errors." % (count))

    if (count > 0):
        for (path, lines) in total_lines:
            print("\nStyle Errors for '%s':" % (path))
            print('---')
            print("\n".join(lines))
            print("---\n")

    return count

def _load_args():
    parser = argparse.ArgumentParser(description =
        "Check the style of all '.py' and '.ipynb' files in the paths specificed (recursivley).")

    parser.add_argument('paths', metavar = 'path',
        type = str, nargs = '+',
        help = 'A path to check for style.')

    return parser.parse_args()

def main():
    return run(_load_args())

if (__name__ == '__main__'):
    sys.exit(main())
