# FSEx changes history

## 2024 June 28
* CLI switch `-1` instructs `fsex` to exit after one command execution. 

## 2024 February 26

CLI flag added:
* `-x {glob}` instructs `fsex` to ignore matching files or directories.  
For the moment only names of files or directories are checked.

## 2023 December 5

Add CLI flags for tuning output:
* `-q` suppresses `fsex` messages
* `-O` suppresses STDOUT of the command executed
* `-E` suppresses STDERR of the command executed

## 2023 August 13

Add `-version` and `-about` flags for checking app details.

## 2023 August 11

In cases when target to monitor (set with `-f` flag) is a directory,
`fsex` will monitor its subdirectories as well.
If new subdirectories are created, they also would be watched.