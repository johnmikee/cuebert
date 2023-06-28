import os
import platform


from util import is_present, shell_out, TOP_LEVEL

CUE_ITEMS = {
    "-h": "cue-default.png",
    "db": "cue-db-usage.png",
    "db info": "cue-db-info.png",
    "db init": "cue-db-init.png",
    "db init schema": "cue-db-init-schema.png",
    "db init values": "cue-db-init-values.png",
    "devices": "cue-devices.png",
    "devices add": "cue-devices-add.png",
    "devices add some -h": "cue-devices-add-some.png",
    "devices get": "cue-devices-get.png",
    "devices get db -h": "cue-devices-get-db.png",
    "devices get mdm -h": "cue-devices-get-mdm.png",
    "devices remove -h": "cue-devices-remove.png",
    "devices update -h": "cue-devices-update.png",
    "db test-connection -h": "cue-db-test-connection.png",
    "exclusions": "cue-exclusions.png",
    "exclusions add -h": "cue-exclusions-add.png",
    "exclusions get -h": "cue-exclusions-get.png",
    "exclusions remove -h": "cue-exclusions-remove.png",
    "users": "cue-users.png",
    "users add": "cue-users-add.png",
    "users add some -h": "cue-users-add-some.png",
    "users get": "cue-users-get.png",
    "users get db -h": "cue-users-get-db.png",
    "users get mdm -h": "cue-users-get-mdm.png",
    "users remove -h": "cue-users-remove.png",
}

CUEBERT_ITEMS = {
    "-h": "cuebert-default.png",
}

OUT_DIR = f"{TOP_LEVEL}/.docs/images"


def get_os() -> str:
    return platform.system().lower()


def main():
    if not is_present("termshot"):
        print(
            "please install termshot (https://github.com/homeport/termshot) and try again"
        )
        os._exit(1)

    os.chdir(TOP_LEVEL)

    for flag, out in CUE_ITEMS.items():
        shell_out(
            [
                "termshot",
                "--show-cmd",
                "--filename",
                f"{OUT_DIR}/{out}",
                "--",
                f"./build/{get_os()}/cue {flag}",
            ]
        )

    for flag, out in CUEBERT_ITEMS.items():
        shell_out(
            [
                "termshot",
                "--show-cmd",
                "--filename",
                f"{OUT_DIR}/{out}",
                "--",
                f"./build/{get_os()}/cuebert {flag}",
            ]
        )


if __name__ == "__main__":
    main()
