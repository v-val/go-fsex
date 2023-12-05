# FSEx changes history

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