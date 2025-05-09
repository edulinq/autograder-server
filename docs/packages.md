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
 5. lockmanager
 6. common, systemserver, stats, jobmanager
 7. email
 8. docker
 9. model
 10. db
 11. analysis, grader, procedures/backup, procedures/logs
 12. procedures/users, report
 13. lms
 14. procedures/courses
 15. scoring
 16. tasks
 17. api, procedures/server
 18. cmd
 19. cmd (non-internal)
