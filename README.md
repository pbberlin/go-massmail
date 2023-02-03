# go-massmail

* Headless massmailing software - ready for cron job automation

* Everything configured in `config.json`

* Email templates and attachments from file system

* Recipient lists in CSV files

* Everything multi-language

* Early beta

## Features

* Read recipient data from CSV

* Compose dynamic messages using templates
  * dynamic field contents
  * file attachments
  * email composition and submission using [go-mail](https://github.com/pbberlin/go-mail)

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

* Due tasks are executed in test mode 24 hours in advance

* Sending via cron job

## Verbose explanations

### Structure

A `project/survey` contains a set of emails.

`Sets` of emails can be sent in recurrent `waves`. A `wave` belongs to a project. It usually has a characteristic month or quarter or season. And a wave has common data points, to which multiple email task can refer.

Each task `task` may have additional data points, for example text elements or attachments.

### Two intricacies

* Tasks with distinct recipients list, while everything else is equal. We could just repeat the config settings, instead we write `SameAs=[otherTask]`. For exactly these case, we also set `TemplateName=xx`, maing it distinct from the default, which would be task name, which would necessary be different for source task and `sameAs` task.

* Each task can be sent via a different SMTP relay host. Since this only rarely needed, there is a default setting for the SMTP relay host.

### Test mode

-mode=test will only send one email for each entry in config TestRecipients.

If the email has more than one language version, test emails are sent for each language and TestRecipient.

If the regular recipient list _contains_ one of the test emails, then this recipient record is chosen for test email.

### Time control

Each task has an execution time.

If program _runtime_ is greater than task execution time,  
but lighter than execution time plus 24h,  
then the task is executed.

24 hours in advance, test emails will be sent for a due task. 

The software is thus intended to be started every day around 10:30 am by cron job.

## Todo

* ReplyTo and Bounce are still unclear.  
  Exchange server bounces are sent to ReplyTo;  
  not to Bounce

* XML example for windows cron

* Batch file setting up logging to file

* The CSV files containing the recipient emails and meta data  
  should be made downloadable via HTTPS request immediately before the task execution.

* Allow inclusion of repeating text blocks (i.e. footer)

* Make functions computing dynamic template fields configurable;  
  at the moment `SetDerived` switches depending on `r.SourceTable` 

* HTML email
  * Outlook is stripping CSS block formatting (float:left etc.)

* HTML inline pictures
  * Inline pictures are not shown by gmail.com
  * [Content type - nested](stackoverflow.com/questions/6706891/)

```log
  Content-Type: multipart/related; boundary="a1b2c3d4e3f2g1"
  --a1b2c3d4e3f2g1
  ...
```

Better

```log
[Headers]
Content-type:multipart/mixed; boundary="boundary1"
--boundary1
Content-type:multipart/alternative; boundary="boundary2"
--boundary2
Content-Type: text/html; charset=ISO-8859-15
Content-Transfer-Encoding: 7bit
[HTML code with a href="cid:..."]

--boundary2
Content-Type: image/png;
name="moz-screenshot.png"
Content-Transfer-Encoding: base64
Content-ID: <part1.06090408.01060107>
Content-Disposition: inline; filename="moz-screenshot.png"
[base64 image data here]

--boundary2--
--boundary1--```

  * Or embedding `<img src="data:image/jpg;base64,{{base64-data-string here}}" />`  
    but "data URIs in emails aren't supported"



* Continue after 8 seconds or keyboard input:  
  can we get rid of the enter key?

