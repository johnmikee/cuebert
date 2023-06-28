# Where should we write the GIF?
Output connect_pg_docker.gif

# Set up the dimensions
Set FontSize 20
Set Width 1200
Set Height 400

# Build the container
Type "make run-docker-postgres"

# Simulate hitting enter
Enter

Sleep 2

# connect to the container
Type "docker exec -it pgcue-db-container bash"

# Simulate hitting enter
Enter

Sleep 2

# login
Type "psql -U cue"

# Simulate hitting enter
Enter

Sleep 2

# show a table
Type "select * from devices;"

# Simulate hitting enter
Enter

Sleep 4
