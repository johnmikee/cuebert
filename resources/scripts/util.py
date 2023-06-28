import subprocess


def is_present(item: str) -> bool:
    from distutils.spawn import find_executable

    return find_executable(item) is not None


def shell_out(cmd: list) -> str:
    return subprocess.check_output(cmd).decode("utf-8").strip()


TOP_LEVEL = shell_out(["git", "rev-parse", "--show-toplevel"])
