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

## Verbose explanation

### Structure

A `project/survey` contains a set of emails.

`Sets` of emails can be sent in recurrent `waves`. A `wave` belongs to a project. It usually has a characteristic month or quarter or season. And a wave has common data points, to which multiple email task can refer.

Each `task` may have additional data points, for example text elements or attachments.

### Relay hosts

* There is a global default setting for the SMTP relay host.

* Each task can have a specific SMTP relay host, if desired

* `DomainsToRelayHorsts` specifies exceptions for certain recipient.  
  For example, internal recipient can be configured to be messaged via the internal SMTP host

* InternalGateway() sniffs, which gateway the sender is connected to.  
  Additional logic can be applied based on the gateway (still work in progress).

* Passwords for relay host auth must be supplied via environment variable.  
  For instance `zimbra.zew.de` would require an advance `export PW_ZIMBRAZEWDE=secret`.

### DRY - dont repeat tasks configurations

* We have tasks with distinct recipients list, while everything else is equal. We could just repeat the config settings, instead we write `SameAs=[otherTask]`.

* If we also wish the same template, we can set `TemplateName=[sourceTask]`.

### Test mode

`-mode=test` will only send one email for each entry in config `TestRecipients`.

If the email has more than one language version, test emails are sent for each language and TestRecipient.

If the regular recipient list and `TestRecipients` _overlap_, then these records are used for the test run.

### Time control

Each task has an execution time.

If program _runtime_ is greater than task execution time,  
but lighter than execution time plus 24h,  
then the task is executed.

24 hours in advance, test emails will be sent for a due task. 

The software is thus intended to be started every day around 10:30 am by cron job.

### URL for CSV files

The CSV files containing the recipient emails and meta data  
can be downloaded via HTTP before the task execution.  

Configurable TTL to enforce ultra fresh recipient lists if need be.

HTTP base64 auth via User setting.  
Password is taken from ENV.

Example URLs

* [FMT invitations](http://fmt-2020.zew.local/fmt/individualbericht-curl.php?mode=invitation)  
* [FMT reminders](http://fmt-2020.zew.local/fmt/individualbericht-curl.php?mode=reminder)  


## Todo

* isInternalGateway():  
   IP addresses need to be configurable  
     map[string]bytes positive  
     map[string]bytes negative  
  isInternalGateway could just be permission or not?  
  Or RelayHorsts could be extended by a "sending-location" key;  
  Thus relay hosts could be selected depending on sender and recipient domain.  
  I need to simplify this.
  
* ReplyTo and Bounce are still unclear.  
  Exchange server bounces are sent to ReplyTo;  
  not to Bounce

* XML example for windows cron

* Batch file setting up logging to file

* Includes of repeating text blocks (i.e. footer)

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
--boundary1--

```

  * Or embedding `<img src="data:image/jpg;base64,{{base64-data-string here}}" />`  
    but "data URIs in emails aren't supported"



* Continue after 8 seconds or keyboard input:  
  can we get rid of the enter key?


## Using MS word and Outlook

This is a fallback in case of an utter emergency.

<https://support.microsoft.com/de-de/office/f%C3%BCr-seriendrucke-verwendbare-datenquellen-9de322a6-f0f9-448d-a113-5fab317d9ef4>

<https://support.microsoft.com/de-de/office/verwenden-des-seriendrucks-zum-senden-von-massen-e-mails-0f123521-20ce-4aa8-8b62-ac211dedefa4>


Empfänger auswählen
Vorhandene Liste auswählen

c:\Users\pbu\Documents\zew_work\git\go\go-massmail\csv\fmt\report-b.csv