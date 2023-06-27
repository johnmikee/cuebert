# Where should we write the GIF?
Output auth_pg_cue.gif

# Set up the dimensions
Set FontSize 20
Set Width 2200
Set Height 400

Type "./build/darwin/cue db init values -users -log-level=info"

Enter

Sleep 1

Type "localhost"

Enter

Sleep 1

Type "cue"

Enter

Sleep 1

Type "cue"

Enter

Sleep 1

Type "5432"

Enter

Sleep 1

Type "cue"

Enter

Sleep 5

Type "./build/darwin/cue db init values -devices -log-level=info"

Enter

Sleep 5

Type "./build/darwin/cue devices get db -all"

Enter

Sleep 2
