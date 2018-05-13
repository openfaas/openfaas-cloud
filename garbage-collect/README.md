## garbage-collect

This function takes a Git owner / repo and list of functions.

The function queries the functions for the owner using the list-functions function - parses the result and then reconciles the differences by deleting any functions which are not in the deployment list but that also match the repo.

### Scenario 1:

Event: initial push

owner: alexellis
repo: alexa-skill
functions: fn1, fn2

Event: second push with fn2 removed

owner: alexellis
repo: alexa-skill
functions: fn1

fn2 is now orphaned so the garbage-collect will remove it.


Event: third push with fn1 renamed to fn3

owner: alexellis
repo: alexa-skill
functions: fn3

fn1 is now orphaned so will be deleted
