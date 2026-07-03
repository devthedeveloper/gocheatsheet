# Runnable Go programs — real backend tasks, real infra

Twelve tiny programs, one concept each — no simulators. Real SMTP (via
MailHog, a drop-in for SendGrid's SMTP relay), real MySQL queries,
a real slow query genuinely cancelled by context. Companion to
[the runnable sheet](https://devthedeveloper.github.io/gocheatsheet/programs.html).

```sh
docker run -d --name mailhog -p 1025:1025 -p 8025:8025 mailhog/mailhog
docker run -d --name mysql-demo -e MYSQL_ROOT_PASSWORD=secret \
  -e MYSQL_DATABASE=notes -p 3306:3306 mysql:8

go run ./01-goroutines     # then check http://localhost:8025
go run -race ./03-mutex    # extra credit: watch Go catch the race
```

`01-goroutines` sends real email over SMTP to MailHog — swap
`localhost:1025` for `smtp.sendgrid.net:587` plus SendGrid API-key auth
and the same code runs against production SendGrid.

Each folder's `output.txt` is the captured output of a real run.
