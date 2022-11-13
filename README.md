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

* Each `project` and each `task` has distinct email templates  
  and distinct attachments

* Each task can be routed to a different SMTP server;  
  for internal or external recipients. 

* Sending via cron job

## Todo

* Allow inclusion of repeating text blocks (i.e. footer)

* Proper separation of function for dynamic template fields by project

* Default to demo mode; command line flags for dry runs and production runs
