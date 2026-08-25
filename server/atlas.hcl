env "local" {
  src = "file://migrations"

  dev = "postgres://postgres:postgres@localhost:5432/leamout_dev?sslmode=disable"

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
