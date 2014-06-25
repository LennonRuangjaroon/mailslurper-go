#
# Use this script to quickly send a bunch of mails. Useful for testing.
#
import sys
import json
import time
import smtplib
import urllib2
import datetime

from email.mime.text import MIMEText
from email.mime.base import MIMEBase
from email.mime.multipart import MIMEMultipart
from email import Encoders

def getQuote():
	response = urllib2.urlopen("http://www.iheartquotes.com/api/v1/random?format=json")
	obj = json.loads(response.read())

	quoteLines = obj["quote"].split("--")

	if len(quoteLines) > 0:
		return {
			"quote": quoteLines[0].strip(),
			"source": "Unknown" if len(quoteLines) <= 1 else quoteLines[1].strip(),
		}
	else:
		return {
			"quote": "No quote",
			"source": "Adam Presley"
		}

if __name__ == "__main__":
	sendMultipartMails = True
	sendTextOnlyMails = True
	sendAttachmentMails = True

	numMails = 5
	address = "127.0.0.1"
	smtpPort = 8000

	me = "someone@another.com"
	to = "bob@bobtestingmailslurper.com"

	try:
		#
		# Send text+html emails
		#
		if sendMultipartMails:
			for index in range(numMails):
				quote = getQuote()

				textBody = "Hello,\nHere is today's quote.\n\n{0}\n  -- {1}\n\nSincerely,\nAdam Presley".format(quote["quote"], quote["source"])
				htmlBody = "<p>Hello,</p><p>Here is today's quote.</p><p><em>{0}</em><br />&nbsp;&nbsp;-- {1}</p><p>Sincerely,<br />Adam Presley</p>".format(quote["quote"], quote["source"],)

				text = MIMEText(textBody, "plain")
				html = MIMEText(htmlBody, "html")

				msg = MIMEMultipart("alternative")

				msg["Subject"] = "Quote From {0}".format(quote["source"])
				msg["From"] = me
				msg["To"] = to
				msg["Date"] = datetime.datetime.now().strftime("%a, %d %b %Y %H:%M:%S +0000 UTC")

				msg.attach(text)
				msg.attach(html)

				server = smtplib.SMTP("{0}:{1}".format(address, smtpPort))
				server.sendmail(me, [to], msg.as_string())
				server.quit()

				#time.sleep(2)

		#
		# Send plain text emails
		#
		if sendTextOnlyMails:
			for index in range(numMails):
				textBody = "Hello,\nI am plain text mail #{0}.\n\nSincerely,\nAdam Presley".format(index,)

				msg = MIMEText(textBody)

				msg["Subject"] = "Text Mail #{0}".format(index,)
				msg["From"] = me
				msg["To"] = to
				msg["Date"] = datetime.datetime.now().strftime("%a, %d %b %Y %H:%M:%S -0700 (UTC)")

				server = smtplib.SMTP("{0}:{1}".format(address, smtpPort))
				server.sendmail(me, [to], msg.as_string())
				server.quit()

				#time.sleep(1)

		#
		# Send text+attachment
		#
		if sendAttachmentMails:
			textBody = "Hello,\nI am plain text mail with an attachment.\n\nSincerely,\nAdam Presley"

			msg = MIMEMultipart()

			msg["Subject"] = "Text+Attachment Mail"
			msg["From"] = me
			msg["To"] = to
			msg["Date"] = datetime.datetime.now().strftime("%a, %d %b %Y %H:%M:%S -0700 (UTC)")

			msg.attach(MIMEText(textBody))

			part = MIMEBase("multipart", "mixed")
			part.set_payload(open("./screenshot.png", "rb").read())
			Encoders.encode_base64(part)
			part.add_header("Content-Type", "image/png")
			part.add_header("Content-Disposition", "attachment; filename=\"screenshot.png\"")
			msg.attach(part)

			server = smtplib.SMTP("{0}:{1}".format(address, smtpPort))
			server.sendmail(me, [to], msg.as_string())
			server.quit()

			#time.sleep(1)

			#
			# Send html+attachment
			#
			htmlBody = "<p>This is a <strong>HTML</strong> email with an attachment.</p>"

			msg = MIMEMultipart()
			html = MIMEText(htmlBody, "html")

			msg["Subject"] = "HTML+Attachment Mail"
			msg["From"] = me
			msg["To"] = to
			msg["Date"] = datetime.datetime.now().strftime("%a, %d %b %Y %H:%M:%S -0700 (UTC)")

			msg.attach(html)

			part = MIMEBase("multipart", "mixed")
			part.set_payload(open("./screenshot.png", "rb").read())
			Encoders.encode_base64(part)
			part.add_header("Content-Type", "image/png")
			part.add_header("Content-Disposition", "attachment; filename=\"screenshot1.png\"")
			msg.attach(part)

			part = MIMEBase("multipart", "mixed")
			part.set_payload(open("./screenshot.png", "rb").read())
			Encoders.encode_base64(part)
			part.add_header("Content-Type", "image/png")
			part.add_header("Content-Disposition", "attachment; filename=\"screenshot2.png\"")
			msg.attach(part)

			server = smtplib.SMTP("{0}:{1}".format(address, smtpPort))
			server.sendmail(me, [to], msg.as_string())
			server.quit()

			#time.sleep(1)

			#
			# Send html+attachment with filename in content-type as "name"
			#
			htmlBody = "<p>This is a <strong>HTML</strong> email with an attachment done differently.</p>"

			msg = MIMEMultipart()
			html = MIMEText(htmlBody, "html")

			msg["Subject"] = "HTML+Attachment Mail 2"
			msg["From"] = me
			msg["To"] = to
			msg["Date"] = datetime.datetime.now().strftime("%a, %d %b %Y %H:%M:%S -0700 (UTC)")

			msg.attach(html)

			part = MIMEBase("multipart", "mixed")
			part.set_payload(open("./screenshot.png", "rb").read())
			Encoders.encode_base64(part)
			part.add_header("Content-Type", "image/png; name=screenshot1.png")
			part.add_header("Content-Disposition", "attachment;")
			msg.attach(part)

			server = smtplib.SMTP("{0}:{1}".format(address, smtpPort))
			server.sendmail(me, [to], msg.as_string())
			server.quit()



	except Exception as e:
		print("An error occurred while trying to connect and send the email: {0}".format(e.message))
		print(sys.exc_info())
