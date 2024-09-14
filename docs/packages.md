# Packages

## Hierarchy

Here is a rough sorting of the packages by dependencies.
Higher numbers depend on lower numbers,
and packages with the same number do not depend on each other.
Most subpackages have not been listed as they either share their parent's level,
or they are a shim/header/interface.
All these packages (except cmd) live in `internal`.

 1. timestamp
 2. log
 3. util
 4. config
 5. common, email
 6. docker
 7. model
 8. db
 9. grader, procedures/logs
 10. procedures/users, report
 11. lms
 12. scoring
 13. task
 14. procedures/courses
 15. api
 16. cmd
