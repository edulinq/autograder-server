# Autograder Assignments

## LMS Syncing Mechanics

When syncing assignment information from an LMS,
only fields not set in the assignment config will be brought over from the LMS.

Matching autograder assignments with LMS assignments are first done via the LMS id set in the assignment's config.
If no LMS is is set, then an attempt is made to match assignments via their name (if the name is not empty).
A name match is made only if an autograder assignment matches one and only one LMS assignment.

## Duplicate Assignments

Assignments in the same course may not share the same ID, name, or LMS ID.
