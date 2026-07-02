# Runnable Go programs

Twelve tiny programs, one concept each. Companion to
[the runnable sheet](https://devthedeveloper.github.io/gocheatsheet/programs.html).

```sh
go run ./01-goroutines
go run -race ./03-mutex   # watch Go catch the data race
```

`11-mysql` needs a database:

```sh
docker run -d --name mysql-demo \
  -e MYSQL_ROOT_PASSWORD=secret -e MYSQL_DATABASE=notes \
  -p 3306:3306 mysql:8
```

Each folder's `output.txt` is the captured output of a real run.
