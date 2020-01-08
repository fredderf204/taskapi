# Task API

This is simple CRUD API to connect to a MongoDB backend, to manage Tasks.

## Requirements

The following env's are required to run the container;

* "AZURE_DATABASE": "aaa",
* "AZURE_DATABASE_PASSWORD": "bbb",
* "AZURE_DATABASE_USERNAME": "ccc",
* "AZURE_DATABASE_HOST": "ddd",
* "GIN_MODE": "debug",
* "PORT": "8080"

\*\* GIN_MODE should stay in debug mode until ready for production/pre-prod. When ready, change env to "release".

\*\* Port 8080 