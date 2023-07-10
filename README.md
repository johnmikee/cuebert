<div align="center">
<h1 align="center">
<img src='.docs/images/header.jpg'width="500" />
<br>
cuebert
</h1>
</div>


## 📚 Table of Contents

- [📍Overview](#overview)
- [⚙️ Credential Configuration](#credential-configuration)
- [🔄 Flow](#flow)
- [🖼️ Images](#images)
- [💬 Messaging](#messaging)
- [⏰ Reminders](#reminders)
- [🧪 Testing](#testing)
- [🗄️ DB](#db)
- [📚 Resources](#resources)
- [🏎💨 Getting Started](#getting-started)
- [🗺 Roadmap](#future-development)
- [🤝 Contributing](#contributing)
- [👏 Acknowledgments](#acknowledgments)

______________________________________________________________________

## Overview

### Cuebert
Cuebert is a slackbot to help encourage users to upgrade the OS on their machine(s). Written in Go this program is intended to be lightweight, fast, and scalable.
Cuebert is simple bot - pass the required OS version and deadline as arguments, and any other relevant flags, and the process will begin. <br />

The goal of Cuebert is empowering the end user. Offering custom reminders and the ability to leverage exclusions, this program seeks to let the user decide when the right time to upgrade is while giving them subtle reminders to do so. By having the fleet on the latest, or latest major, release we reduce a significant attack vector. Apple has made it clear the only fully updated OS is the latest OS. So while old versions might get some security patches they are not getting the full attention of the latest version. <br />

When Apple pushes out a new release IT will test the updates. Once the update has met the acceptance criteria the minimum required version and date will be updated and Cuebert will begin its reminder cycle.
<br />

This program was written using Slack's [socket mode](https://api.slack.com/apis/connections/socket) making testing quick and easy. There is no need for callback urls or public listener's, Cuebert will be "live" wherever you run it from. <br />

______________________________________________________________________

## Credential Configuration
There are three ways to store the required credentials for the program environmental variables, configuration file, and the system keychain.<br />

If testing locally the program will default to `-env-type=dev` which handles authentication through the system keychain using [keyring](https://github.com/zalando/go-keyring). If running the program in production it is recommended to set the credentials as environmental variables.<br />

To set the environmental variables take any of the json key [values](cuebert/config.go#L17-28) in the `Config` struct, prepend them with `CUEBERT` and export the value. <br />

ex: <br />
```
export CUEBERT_DB_ADDRESS=localhost
export CUEBERT_DB_USER=cue
```
<br />

While the other two methods are preferred, you may also store the values in a local json configuration file. <br />
```
{
    "db_address": "localhost",
    "db_name": "cue",
    ...
}
```
<br />

______________________________________________________________________
## Flow
By default when the program starts the four tables are deleted. Pulling the data and populating the tables takes roughly 2-3 seconds and ensures we start with a fresh data set. Should you wish to keep the data pass `-clear-tables=false`. <br />

After reading in the configuration needed to start the first action takin is to start a goroutine to respond to user input. This handles any messages sent to Cuebert, reminder modal input, acknowledgement of messages sent, or any information Cuebert would need to respond to. <br />

Once the configuration has been read in and the tables have been populated the `bot_results` table is configured. This pulls data from the `devices` and `users` table to gather the information on the user, and their manager, needed to begin.<br />
<br />

At this point Cuebert begins maintenance tasks. Three routines are started and will repeat on a set time interval.<br />

* Device Diff
    - Every 60 minutes the [deviceDiff](cuebert/devices.go) function runs to pull info from the DB, compare that to the MDM information, and update where necessary. This ensures we are not messaging users who have updated in between reminders and always operating of fresh device data. If a device is above the required version it will be dropped from the table during this.
<br />

* Check
    - Every 30 minutes [check](cuebert/check.go) runs to see if messaging needs to be sent to the user. The information is pulled from the `bot_results` table which stores if a user has been sent the message and when/if they responded. If the conditions are met where the user should be notified that will occur.
<br />

* Poll Reminders
    - Every 15 minutes [pollReminders](cuebert/check.go) runs to see if anyone needs a custom OS upgrade reminder set. Since a user can set a reminder for any point, at any time, we need to regularly check this data. Should a reminder be set in within the next 15 minutes of check running a routine is started to remind the user at that time.
<br />
______________________________________________________________________

## Messaging 
There are two main messages Cuebert will send. The first is sent once a new os requirement is set to inform the user an upgrade is required.<br />
![alt text](.docs/images/first_message.png)
<br />
This notification comes with a button for the user to acknowledge they have received the message. When the button is clicked it is recorded that the message has been read and the time at which this was done. This action allows us to verify that the user knows there is a required update and that they have agreed to do it within the alloted time frame.<br />

The second message is sent midweek which creates a group DM between Cuebert,the users manager, and the user _if_ the device is still out of compliance.
______________________________________________________________________

## Reminders
At anytime a user can request a reminder to upgrade. A prompt appears to ask if the user would like to set a reminder.<br />
![alt text](.docs/images/want_reminder.png)
<br />

If the user clicks yes a date/time modal appears to set the reminder.<br />
<br />
![alt text](.docs/images/date_time_modal.png)
<br />
______________________________________________________________________

## Testing
By passing cuebert the `-testing` and `-testing-users` flags you can simulate the actions that would take place during messaging.<br />

The testing flag is a bool value. The testing users is a comma-separated lists of Slack ID's to act on as the program normally would while logging what would happen to others. This flag is very useful for beta testing against IT and other willing participants.

An example of a possible testing scenario:
> `./build/darwin/cuebert -deadline-date 2023-05-09 -cutoff-time 18:00:00 -log-level=trace -required-os 13.4.1 -auth-users=ABC123,DEB456 -testing=true -clear-tables=false -testing-users=ABC123`

## Setup
To test locally we first need to setup Postgres. This can be done by downloading the standalone Postgres [package](https://www.postgresql.org/download/macosx/) or by creating a Postgres Docker image.<br />

The docker image is going to be quicker to get up and running but either option will work.

<br />

### Docker Postgres
`make run-docker-postgres`<br />
This will build the [Dockerfile](resources/Docker/postgres/Dockerfile), configuring the database as we need it and starting Postgres.

To connect to the db you can:
* exec into the container<br />

![alt text](.docs/gifs/connect_pg_docker.gif)
<br />

* use `psql` if you have the postgres binaries installed locally.

![alt text](.docs/gifs/connect_pg_psql.gif)
<br />
<br />

### Standalone Postgres
* Download the [latest](https://www.postgresql.org/download/macosx/) version.
* Move to Applications folder ➜ Open.
* Click `Initialize`<br />
![alt text](.docs/images/pg_local_init.png)
* You now have a PostgreSQL server running on your Mac with these default settings:
    - Host: localhost
    - Port: 5432
    - User: your system user name
    - Database:	same as user
    - Password:	none
* Configure your $PATH to use the included command line tools (optional):
    - `sudo mkdir -p /etc/paths.d && echo /Applications/Postgres.app/Contents/Versions/latest/bin | sudo tee /etc/paths.d/postgresapp`
* Run the init [script](resources/db/create.sh) to configure the DB as needed. <br />
![alt text](.docs/gifs/init_pg_local.gif)

---

Once the database has been configured run cuebert to interact as needed.
* Unless you have changed some of the default values the connection information should:
    - Host: localhost
    - Port: 5432
    - User: cue
    - Database:	cue
    - Password:	cue
<br />

Now you can authenticate and start using the tools.
![alt text](.docs/gifs/auth_pg_cue.gif)
<br />
______________________________________________________________________

## DB
The database stores information on the users, devices, bot results, and exclusions. Each table has its own tooling defined under [db](db/) that help facilitate interactions with the table. By design the functionality and code is very similar between each of these packages - they are meant to be fully independent of one another while feeling similar when moving between them.
<br />

## Tables
The four table definitions can be found [here](resources/db/create_tables.sql)<br />
* bot results<br />
    - This table is the one Cuebert will be writing state information to about interactions with the user such as when a user acknowledges or receives a message, the time it occurred, etc.
* devices<br />
    - Information about the device. All information is pulled from the MDM to store the device serial, os, platform, and user.
* users<br />
    - Used to correlate information between the MDM device users and their Slack ID.
* exclusions<br />
    - Devices to be excluded from receiving messaging.
<br />

### Creating tables
To help automate the creation of the tables, especially when testing, there is a [script](resources/db/create.sh) that can be used.<br />
This does make some assumptions on the default user of the db (postgres) but that can be overridden with flags.
<br />

Create the cue user, db, tables, and triggers.<br />
`bash resources/db/create.sh -a true`<br />
Create the cue user<br />
`bash resources/db/create.sh -u true`<br />
Create the DB<br />
`bash resources/db/create.sh -d true`<br />
Create the tables<br />
`bash resources/db/create.sh -t true`<br />
Create the triggers<br />
`bash resources/db/create.sh -tr true`<br />
<br />
______________________________________________________________________

## Resources
### Images
Under [images](resources/images/) you will find some sample assets you can use for both `Cuebert` and `Bertcue`
<br />

### Manifests
Slack's App [Manifests](https://api.slack.com/reference/manifests) make it easy to get this up and running.<br />
Under [manifest](resources/manifest) there are two options to choose from:
* [prod](resources/manifest/prod.json): this will create the app as `Cuebert`
* [testing](resources/manifest/testing.json): this will create the testing instance `Bertcue`
<br />

Since this application uses socketmode testing with the same app, if already deployed into production, leads to a split-brain problem.<br />
Messages will deliver to whichever app it feels like and will cause a world of pain and confusion. 
<br />

To save yourself from this, simply deploy a second app instance and test away.
______________________________________________________________________

## Getting Started

### Using cuebert

```sh
./build/darwin/cuebert -h      
Usage of ./build/darwin/cuebert:
  -auth-users string
        Set which users can perform authorized functions. (comma separated)
  -auth-users-from-idp
        Set whether to pull authorized users from the IDP. (default true)
  -check-interval int
        the number of minutes between device messaging checks and db clean-ups. (default 15)
  -clear-tables
        Drop all info from tables on initialization. (default true)
  -cutoff-time string
        the hour when the install must be done by (HH:MM:SS).
  -daily-report
        send a daily report to the admin alert channel.
  -deadline-date string
        the date the install must be done by (YYYY:MM:DD).
  -default-reminder-interval int
        the number of minutes between reminders. (default 60)
  -device-diff-interval int
        the number of minutes between device diff checks. (default 30)
  -env-type string
        Set the env type. Options are [prod, dev]. (default "dev")
  -help-docs-url string
        the url to the cuebert docs. (default "https://ramp.slab.com/posts/cuebert-hem79mkb")
  -help-repo-url string
        the url to the cuebert repo. (default "https://github.com/ramp/cue")
  -help-ticket-url string
        the url to the cuebert ticketing system. (default "https://linear.app/tryramp/team/IT/active")
  -idp string
        Set the IDP to use. Options are [okta]. (default "okta")
  -init
        Start the program, load the config, and wait for input before running. (default true)
  -log-level string
        Set the log level. (default "trace")
  -log-to-file
        Log results to file.
  -mdm string
        Set the MDM to use. Options are [jamf, kandji]. (default "kandji")
  -method string
        Set the method to use. Options are [manager, device]. (default "manager")
  -poll-interval int
        the number of minutes between device polling checks. (default 10)
  -rebuild-tables-on-failure
        rebuild tables on an abnormal exit.
  -required-os string
        the version to require for the fleet (default "13.4.1")
  -send-manager-missing
        send a message to the alert channel of missing managers
  -service-name string
        if using the dev env the service name to store keys under. (default "cuebert")
  -table-names string
        a list of tables to clear on initialization. (comma separated) (default "bot_results,devices,exclusions,users")
  -testing
        Log actions that would take place instead of performing them. (default true)
  -testing-end-time string
        the time to end testing (HH:MM). (default "17:00")
  -testing-start-time string
        the time to start testing (HH:MM). (default "11:00")
  -testing-users string
        a list of slack id's to perform the actions on during testing instead of every user. (comma separated)
```
<hr />

## Future Development

- \[📌\] \[Initial release\]
- \[ \] \[📌  Finish Tests\]
- \[ \] \[📌  Multiple OS Support\]
- \[ \] \[📌  Multiple iDP Support\]
- \[ \] \[📌  Multiple MDM Support\]
- \[ \] \[📌  Zero-day flow for Cuebert\]
- \[ \] \[📌  Creation of Cubes (actions to run if deadline passed)\]
______________________________________________________________________

## Contributing

Contributions are always welcome! Please follow these steps:

1. Clone the repository.
2. Create a new branch with a descriptive name (e.g., `new-feature-branch` or `cuebert-bugfix-123`).
```sh
git checkout -b new-feature-branch
```
3. Make the changes, commit with a clear commit message that explains the changes you've made, and push to the new branch.
```sh
git commit -m 'fixed cuebert going sentient'
git push -u origin cuebert-selfaware-removal
```
4. Create a pull request to the original repository and describe the changes you've made and why they're necessary.


## Acknowledgments
This project was heavily inspired by the work of both [micromdm](https://github.com/micromdm/micromdm) and [nanomdm](https://github.com/micromdm/nanomdm) and anything smart logic wise probably came from there.
