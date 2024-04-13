from email.mime.text import MIMEText
import smtplib
msg = MIMEText('hello, send by Python...', 'plain', 'utf-8')


from_addr = "admin@domain.com"
password = "admin"
to_addr = "admin@domain.com"
smtp_server = "127.0.0.1"

server = smtplib.SMTP(smtp_server, 25)
server.starttls()
server.login(from_addr, password)
server.sendmail(from_addr, [to_addr], msg.as_string())
server.quit()