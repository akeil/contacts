# NAME

Usecase

View details for a single contact.


## Actors
User


## Preconditions
Addressbook exists
Contact exists


## Postconditions
User knows contact details


## Main Course
1. User types `find` with a searchstring argument
2. App retrieves matching contact
3. App displays contact details


## Alternate Courses
When (2) returns more than one match,
show a *list* of matches and let the user select one.


## Comments
