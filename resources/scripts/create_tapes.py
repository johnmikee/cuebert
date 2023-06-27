import argparse
import os
import shutil

from util import is_present, shell_out, TOP_LEVEL


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="tape_converter",
        description=(
            "Will convert any tape file passed to a gif. If no arguments are passed all tape files will be generated."
        ),
        allow_abbrev=False,
    )

    parser.add_argument(
        "--tape",
        type=str,
        help="The name of the tape file to convert",
        required=False,
    )

    parser.add_argument(
        "--output",
        type=str,
        help="The name output file",
        required=False,
    )

    return parser.parse_args()


def iterate_tapes(resource_dir: str) -> list[dict]:
    tapes = []
    for _, _, files in os.walk(resource_dir):
        for file in files:
            tapes.append(
                {"tape": file, "output": f'{file.replace("-", "_").split(".")[0]}.gif'}
            )

    return tapes


def convert(top_level: str, resource_dir: str, tape_args: list[dict]) -> None:
    os.chdir(top_level)

    for tape in tape_args:
        shutil.copy(src=f"{resource_dir}/tapes/{tape['tape']}", dst=top_level)

        shell_out(["vhs", tape["tape"]])

        os.remove(tape["tape"])

        shutil.move(src=tape["output"], dst=f"{resource_dir}/gifs/{tape['output']}")


def main():
    if not is_present("vhs"):
        print(
            "please install vhs (https://github.com/charmbracelet/vhs) and try again."
        )
        os._exit(1)

    args = parse_args()

    tape_args = []
    resource_dir = f"{TOP_LEVEL}/.docs"

    if not (args.tape and args.output):
        tape_args = iterate_tapes(resource_dir=f"{resource_dir}/tapes")

    else:
        output = args.output
        if not output.endswith(".gif"):
            output = f"{output}.gif"

        tape_args = [{"tape": args.tape, "output": output}]

    convert(top_level=TOP_LEVEL, resource_dir=resource_dir, tape_args=tape_args)


if __name__ == "__main__":
    main()
