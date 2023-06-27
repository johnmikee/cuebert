# Where should we write the GIF?
Output init_pg_local.gif

# Set up the dimensions
Set FontSize 20
Set Width 1200
Set Height 400

# Build the container
Type "bash resources/db/create.sh -a true -n cue"

# Simulate hitting enter
Enter

Sleep 2

# connect to the container
Type "psql --host=localhost --user=cue"

# Simulate hitting enter
Enter

Sleep 2

# show a table
Type "select * from users;"

# Simulate hitting enter
Enter

Sleep 4
