# go-massmail

* Early alpha

* Read recipient data from CSV

* Compose dynamic messages using templates
  * dynamic functions
  * file attachments
  * email structures and submission using [go-mail](https://github.com/pbberlin/go-mail)

* Each `project/survey` can have multiple `waves`,  
  representing monthly or quarterly repetitions

* Each `project/survey` can have multiple `tasks`,  
  representing for instance `invitations`, `reminders` and `results`

* Each `project/survey` and each `task` has distinct email templates  
  and distinct attachments

* Each task can be routed to a different SMTP server;  
  for internal or external recipients. 

* Command line flag `mode=[test|prod]`  
  for test and production runs

* Sending via cron job

## Structure

A `project/survey` contains a set of emails.

`Sets` of emails can be sent in recurrent `waves`. A `wave` belongs to a project. It usually has a characteristic month or quarter or season. And a wave has common data points, to which multiple email task can refer.

Each task `task` may have additional data points, for exanmple text elements or attachments.

### Two intricacies

* Tasks with distinct recipients list, while everything else is equal. We could just repeat the config settings, instead we write `SameAs=[otherTask]`. For exactly these case, we also set `TemplateName=xx`, maing it distinct from the default, which would be task name, which would necessary be different for source task and `sameAs` task.

* Each task can be sent via a different SMTP relay host. Since this only rarely needed, there is a default setting for the SMTP relay host.

## Test mode

-mode=test will only send one email for each entry in config TestRecipients.

If the email has more than one language version, test emails are sent for each language and TestRecipient.

## Time control

Each task has an execution time.

If program _runtime_ is greater than task execution time,  
but lighter than execution time plus 24h,  
then the task is executed.

24 hours in advance, test emails will be sent for a due task. 

The software is thus intended to be started every day around 10:30 am by cron job.

## Todo

* The CSV files containing the recipient emails and meta data,  
  should be made downloadable via HTTPS request immediately before the task execution.

* Allow inclusion of repeating text blocks (i.e. footer)

* Proper separation of function for dynamic template fields by project

