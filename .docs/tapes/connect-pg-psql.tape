# Where should we write the GIF?
Output connect_pg_psql.gif

# Set up the dimensions
Set FontSize 20
Set Width 1200
Set Height 400

# Build the container
Type "make run-docker-postgres"

# Simulate hitting enter
Enter

Sleep 5

# connect to the container
Type "psql --host=localhost --user=cue"

# Simulate hitting enter
Enter

Sleep 2

# login
Type "cue"

# Simulate hitting enter
Enter

Sleep 2

# show a table
Type "select * from devices;"

# Simulate hitting enter
Enter

Sleep 4
