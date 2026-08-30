env "local" {
  url = getenv("DATABASE_URL")
  dev = getenv("ATLAS_DEV_DATABASE_URL")

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "deploy" {
  url = getenv("DATABASE_URL")

  migration {
    dir = "file://migrations"
  }
}
