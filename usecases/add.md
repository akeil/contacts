# NAME

Usecase

Add a new contact to an addressbook.


## Actors
User


## Preconditions
Addressbook exists.


## Postconditions
New contact exists in addressbook


## Main Course
1. User types  `add`
2. App opens *Contact Template* with all fields empty in default editor
3. User edits template
4. User saves template and closes editor
5. App parses values from template
6. creates new contact


## Alternate Courses
In (1), user supplies contact details as command line args. App displays these values in the template.

In (1), user provides contact details as command line args *and* an additional flag to not edit the contact.
We skips (2) through (5) and go to (6)


## Comments
editor to be used can be set in config
editor template can be set in config
