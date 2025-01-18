# Packages

## Hierarchy

Here is a rough sorting of the packages by dependencies.
Higher numbers depend on lower numbers,
and packages with the same number do not depend on each other.
Most subpackages have not been listed as they either share their parent's level,
or they are a shim/header/interface.
All these packages (except cmd) live in `internal`.

 1. exit, timestamp
 2. log
 3. util
 4. config
 5. common, stats
 6. email
 7. docker
 8. model
 9. db
 10. analysis, grader, procedures/backup, procedures/logs
 11. procedures/users, report
 12. lms
 13. procedures/courses
 14. scoring
 15. task
 16. api
 17. cmd
