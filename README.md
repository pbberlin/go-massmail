# go-massmail

* Early alpha

* Read recipient data from CSV

* Compose dynamic messages using templates
  * dynamic functions
  * file attachments
  * email structures and submission using [go-mail](https://github.com/pbberlin/go-mail)

* Each `project` can have multiple `waves`,  
  representing monthly or quarterly repetitions

* Each `project` can have multiple `tasks`,  
  representing for instance `invitations`, `reminders` and `results`

* Each `project` and each `task` has distinc email templates  
  and distinct attachments

* Sending via cron job

## Todo

* Allow inclusion of repeating text blocks (i.e. footer)

* Proper separation of function for dynamic template fields by project