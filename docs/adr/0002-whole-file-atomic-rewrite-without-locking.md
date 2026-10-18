# Whole-file atomic rewrite, without locking

Every command loads the entire task file, changes it in memory and writes it back by writing a temporary file in the same directory and renaming it over the original, so a crash mid-write leaves either the old or the new file, never a half-written one. A task file that fails to parse is reported as an error and never overwritten. We deliberately do not lock the file: two `task-cli` processes running at the same moment can lose one of the two changes, which is acceptable for a single-user CLI.
