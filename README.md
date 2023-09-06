# go-massmail

* Headless massmailing software - ready for cron job automation

* Everything configured in `config.json`

* Email templates from file system

* Attachments from file system  
  * TTL to prevent mistakes

* Recipient lists in CSV files    
  * updated via `wget` with TTL

* Dynamic string computation for reminder dates, salutations, etc.

* Everything multi-language

* Early beta

## Features

* Read recipient data from CSV

* Compose dynamic messages using templates
  * dynamic field contents
  * file attachments
  * email composition and submission using [mailyak](https://github.com/domodwyer/mailyak)

* Each `project/survey` can have multiple `waves`,  
  representing monthly or quarterly repetitions

* Each `project/survey` can have multiple `tasks`,  
  representing for instance `invitations`, `reminders` and `results`

* Each `project/survey` and each `task` has distinct email templates  
  and distinct attachments

* Each task can be routed to a different SMTP server;  
  different SMTP server, based on recipients' domain. 

* Command line flag `-mode=[test|prod]`  
  for test and production runs  
  integration of DKIM testers i.e. `mail-tester.com` or `mxtoolbox.com`

* Command line flag `-start=[2023-09-06T09:15]`  
  defers running the program until given date and time.  
  Default is now().  
  The first due task is executed in preflight.  
  Then the application goes into wait loop.  
  At the end of the wait loop, the preflight run is repeated.  
  If there are more than one due tasks, only the first one  
  is preflighted before the wait loop.  

* Due tasks are executed in test mode 24 hours in advance

* Sending via cron job

## Verbose explanation

### Structure

A `project/survey` contains a set of emails.

`Sets` of emails can be sent in recurrent `waves`. A `wave` belongs to a project. It usually has a characteristic month or quarter or season. And a wave has common data points, to which multiple email tasks can refer.

Each `task` may have additional data points, for example text elements or attachments.

### Relay hosts

* There is a global default setting for the SMTP relay host.

* Pause between emails distinct for each host.

* Each task can have a specific SMTP relay host, if desired.

* Passwords for relay host auth must be supplied via environment variable.  
  For instance `zimbra.zew.de` would require an advance `export PW_ZIMBRAZEWDE=secret`.

* `DomainsToRelayHorsts` specifies exceptions for certain recipients.  
  For example, if internal recipients can only be reached through the internal SMTP host

#### Sunsetted/frozen feature

* InternalGateway() sniffs, which gateway the sender is connected to.  
  Additional logic for relay host selection could be applied based on the gateway.  

### DRY - dont repeat yourself - tasks configurations

* We may have tasks with distinct recipients list, while everything else is equal.  
  We could just repeat the config settings, instead we write `SameAs=[otherTask]`.

* CSV file name is always derived from the task name.

* Template name derives automatically from the task name.  
  We can set a different template via `TemplateName=[sourceTask]`.

* HTML templates - centered layout with max width.  
  Outlook strippings CSS block formatting (float:left etc.).  
  Max-width can only be done using a table: <stackoverflow.com/questions/2426072/>.  
  See the invitation template of the pds project for an example.  
  The basic structure is:

```html
<table border="0" cellspacing="0" width="100%">
    <tr>
        <td></td>
        <td width="350">350 pixels max, but shrink if less.
        </td>
        <td></td>
     </tr>
</table> 
```


* Partial templates must beging with _partial-[langcode]-_  
  These templates can be embedded into the main template via    
  `{{template "partial-de-footer.html" .}}`  
  Dont create more than a dozen, as each is parsed with every main template.

### Time control

Each task has an `execution time` or an `execution interval`.

1. `execution interval` currently only supports `daily`,  
in effect running the task _every_ time.

2. If program _runtime_ is greater than task `execution time`,  
but lighter than execution time plus 24h,  
then the task is executed.

24 hours in advance, test emails will be sent for a due task. 

The software is thus intended to be started every day around 10:30 am by cron job.

### Test mode

`-mode=test` will only send one email for each entry in config `TestRecipients`.

If the email has more than one language version, test emails are sent for each language and TestRecipient.

If the regular recipient list and `TestRecipients` _overlap_, then these records are used for the test run.

There is still conceptual overlap between explicit test tasks with `ExecutionInterval=daily`.

### URL for CSV files

The CSV files containing the recipient emails and meta data  
can be downloaded via HTTP before task execution.  

Configurable TTL to enforce ultra fresh recipient lists if need be.

HTTP base64 auth via User setting.  
Password is taken from ENV.

Example URLs (internal from ZEW institute)

* [FMT invitations](http://fmt-2020.zew.local/fmt/individualbericht-curl.php?mode=invitation)  
* [FMT reminders](http://fmt-2020.zew.local/fmt/individualbericht-curl.php?mode=reminder)  

`UserIDSkip` is a map of user IDs that should be omitted from the CSV.  
This is a quick and dirty way to send reminders to those recipients,  
who have not yet answered.

## Todo

* ReplyTo and Bounce (via header `"Return-Path"`) are still unclear.      
  Exchange server bounces are sent to ReplyTo;  
  not to Bounce.

* XML example file for windows cron

* Batch file, setting up logging to file

### Todo templates

* HTML email templates should get a distinct plain text version.  
  At the moment, we just add the HTML file again as plain text.

* Make functions computing dynamic template fields configurable;  
  at the moment `SetDerived` switches depending on `r.SourceTable` etc. 


### Todo Prio C

* Continue after 8 seconds or keyboard input:  
  can we get rid of the enter key?

* isInternalGateway() - should we drop this? :  
   IP addresses need to be configurable  
     map[string]bytes positive  
     map[string]bytes negative  
  isInternalGateway could just be permission or not?  
  Or RelayHorsts could be extended by a "sending-location" key;  
  Thus relay hosts could be selected depending on sender and recipient domain.  
  I need to simplify this.
  

### Old considerations - if going back to go-mail

* HTML inline pictures issue  
  was solved by switching to `github.com/domodwyer/mailyak`.  

* We could extend go-mail  
  [Content type - nested](stackoverflow.com/questions/6706891/)

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
--boundary1--

```

* Or embedding `<img src="data:image/jpg;base64,{{base64-data-string here}}" />`  
  but "data URIs in emails aren't supported"


## Using MS word and Outlook

A fallback in case of an extreme emergency.

<https://support.microsoft.com/de-de/office/f%C3%BCr-seriendrucke-verwendbare-datenquellen-9de322a6-f0f9-448d-a113-5fab317d9ef4>

<https://support.microsoft.com/de-de/office/verwenden-des-seriendrucks-zum-senden-von-massen-e-mails-0f123521-20ce-4aa8-8b62-ac211dedefa4>

German menu items
* `Empfänger auswählen`
* `Vorhandene Liste auswählen`
* Choose file  
   c:\Users\pbu\Documents\zew_work\git\go\go-massmail\csv\fmt\report-b.csv