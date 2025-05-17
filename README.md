# Simple Scheduler

[![Build and Test Status](https://github.com/jacobmcgowan/simple-scheduler/actions/workflows/go-build-test.yml/badge.svg)](https://github.com/jacobmcgowan/simple-scheduler/actions/workflows/go-build-test.yml)
[![License: MIT](https://cdn.prod.website-files.com/5e0f1144930a8bc8aace526c/65dd9eb5aaca434fac4f1c34_License-MIT-blue.svg)](/LICENSE)

Simple service to schedule jobs to run at specific times or in intervals.

## Services
This monorepo contains multiple services for running and managing Simple
Scheduler. These can be found in the [services](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services) directory.

### Scheduler
This service schedules runs for jobs. Each instance of the Scheduler will lock
the jobs that it manages to prevent other instances from scheduling duplicate
runs.

Execution for runs is handled via a message bus using a pub/sub model. When the
Scheduler starts a run of a job it publishes a message to start the job. It is
assumed that a job worker service will consume this message then perform its
task. The job worker should publish status messages that the Scheduler will
consume to update the run's status.

Currently, RabbitMQ is the only supported message bus but other options may be
added in future updates.

#### Run Actions
The following actions can be published by the Scheduler:

| Action | Description               | Expected Status Responses      |
|--------|---------------------------|--------------------------------|
| run    | The run should be started | running \| failed \| completed |
| cancel | The run should be stopped | cancelled                      |

These actions are published to the `scheduler.job.N.action` exchange where `N`
is the name of the job. The body of the action message is a JSON object in the
following format:

```json
{
    "jobName": "my-job",
    "runId": "6799b53b33fcc6482f29c96f",
    "action": "run"
}
```

#### Run Status
The following statuses are supported for runs:

| Status     | Set By    | Description                                                                          |
|------------|-----------|--------------------------------------------------------------------------------------|
| pending    | Scheduler | Run has been scheduled and the run action has been published                         |
| running    | Client    | The run action has been received and the run has started                             |
| cancelling | Scheduler | The run has been cancelled by the Scheduler and the cancel action has been published |
| cancelled  | Client    | The cancelled action has been received and the run has stopped                       |
| failed     | Client    | The run has failed                                                                   |
| completed  | Client    | The run has finished successfully                                                    |

These statuses are published to the `scheduler.job.N.status` exchange where `N`
is the name of the job. The body of the status message is a JSON object in the
following format:

```json
{
    "jobName": "my-job",
    "runId": "6799b53b33fcc6482f29c96f",
    "status": "running"
}
```

#### Heartbeats
To be able to determine whether a run is still being worked on by a client or if
there has been an unexpected crash or another issue, the Scheduler expects to
receive heartbeat messages from the client. These messages should be published
to the `scheduler.job.N.heartbeat` exchange where `N` is the name of the job.

The body of heartbeat messages are a JSON object in the following format:

```json
{
    "jobName": "my-job",
    "runId": "6799b53b33fcc6482f29c96f"
}
```

If a heartbeat message is not received within the configured heartbeat timeout
for the job, then the run's status will be reset to `pending` and another `run`
action will be published.

#### Adding support for alternative message bus services
To implement support for a different message bus
service, refer to the [MessageBus interface](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services/scheduler/message-bus/message-bus.go).
Add a directory for the new message bus option in [services/scheduler/message-bus](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services/scheduler/message-bus)
and add a constant for it to [MessageBusType](https://github.com/jacobmcgowan/simple-scheduler/tree/main/services/scheduler/message-bus/message-bus-types/message-bus-types.go).
Lastly, update [RegisterMessageBus()](https://github.com/jacobmcgowan/simple-scheduler/blob/main/shared/resources/resources.go#L45)
to support the configuration option to use the new message bus service.

#### Running
The Scheduler currently has the following dependencies:
- [MongoDB](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-community-with-docker/)
- [RabbitMQ](https://www.rabbitmq.com/docs/download)

Once the dependencies are running, you can run the service locally using
```bash
cd services/scheduler
go build .
./scheduler
```

or run in Docker using
```bash
docker build -f ./services/scheduler/Dockerfile -t simple-scheduler/scheduler .
docker run -it --rm --name simple-scheduler -p 8080:8080 simple-scheduler/scheduler:latest
```

#### Settings
The Scheduler supports the following settings set in the .env file:

| Environment Variable                          | Description                                                                                |
|-----------------------------------------------|--------------------------------------------------------------------------------------------|
| SIMPLE_SCHEDULER_MAX_JOBS                     | The maximum number of jobs the scheduler instance will manage. If 0, then no limit is set. |
| SIMPLE_SCHEDULER_DB_TYPE                      | The type of database to use. e.g. `mongodb`.                                               |
| SIMPLE_SCHEDULER_MESSAGEBUS_TYPE              | The type of message bus to use. e.g. `rabbitmq`.                                           |
| SIMPLE_SCHEDULER_DB_CONNECTION_STRING         | The connection string of the database.                                                     |
| SIMPLE_SCHEDULER_DB_NAME                      | The name of the database to connect to.                                                    |
| SIMPLE_SCHEDULER_MESSAGEBUS_CONNECTION_STRING | The connection string of the message bus.                                                  |
| SIMPLE_SCHEDULER_CLEANUP_INTERVAL             | The interval in milliseconds to cleanup stuck runs.                                        |
| SIMPLE_SCHEDULER_CACHE_REFRESH_INTERVAL       | The interval in milliseconds to refresh the job cache.                                     |
| SIMPLE_SCHEDULER_HEARTBEAT_INTERVAL           | The interval in milliseconds to set the heartbeat for locked jobs.                         |

### Custodian
This service cleans up locked jobs in the event that an instance of the
Scheduler service crashes or stops unexpectedly without unlocking the jobs it
was managing.

#### Running
The Custodian currently has the following dependencies:
- [MongoDB](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-community-with-docker/)

Once the dependencies are running, you can run the service locally using
```bash
cd services/custodian
go build .
./custodian
```

or run in Docker using
```bash
docker build -f ./services/custodian/Dockerfile -t simple-scheduler/custodian .
docker run -it --rm --name simple-scheduler-custodian -p 8080:8080 simple-scheduler/custodian:latest
```

#### Settings
The Scheduler supports the following settings set in the .env file:

| Environment Variable                          | Description                                                                                |
|-----------------------------------------------|--------------------------------------------------------------------------------------------|
| SIMPLE_SCHEDULER_DB_TYPE                      | The type of database to use. e.g. `mongodb`.                                               |
| SIMPLE_SCHEDULER_DB_CONNECTION_STRING         | The connection string of the database.                                                     |
| SIMPLE_SCHEDULER_DB_NAME                      | The name of the database to connect to.                                                    |
| SIMPLE_SCHEDULER_CLEANUP_INTERVAL             | The interval in milliseconds to cleanup stuck jobs.                                        |
| SIMPLE_SCHEDULER_HEARTBEAT_TIMEOUT            | The time period in milliseconds to wait for a heartbeat before unlocking a job.            |

### API
This service provides a RESTful API to manage jobs and runs.

#### Running
The API currently has the following dependencies:
- [MongoDB](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-community-with-docker/)
- An OpenID provider such as [Keycloak](https://www.keycloak.org/getting-started/getting-started-docker)

If using Keycloak, you can import the [example realm](examples/keycloak-example-realm.json)
but be aware that this example configuration is for development purposes only
and should not be used in production. This example includes the following:

- Realm: simple-scheduler
- Client:
  - Client ID: simple-scheduler-cli
  - Client Secret: See [Clients > simple-scheduler-cli > Credentials](http://localhost:8080/admin/master/console/#/simple-scheduler/clients/9b9111f9-a0bd-499f-b7e3-9788b18165b2/credentials)
  - Roles:
    - jobs:read
    - jobs:write
    - runs:read
    - runs:write
  - Scopes:
    - jobs:read
    - jobs:write
    - runs:read
    - runs:write
- User:
  - Usernname: guest
  - Password: guest
  - Roles:
    - jobs:read
    - jobs:write
    - runs:read
    - runs:write

If using a different OpenID provider or a different configuration, ensure that
you create a client for the CLI and any other client that will access the API
with a redirection URI of `http://localhost:5556/callback` (used by the CLI) and
the following roles and scopes:
- jobs:read
- jobs:write
- runs:read
- runs:write

Once the dependencies are running, you can run the API locally using
```bash
cd services/api
go build .
./api
```

or run in Docker using
```bash
docker build -f ./services/api/Dockerfile -t simple-scheduler/api .
docker run -it --rm --name simple-scheduler-api -p 8080:8080 simple-scheduler/api:latest
```

#### Settings
The API supports the following settings set in the .env file:

| Environment Variable                          | Description                                                                                |
|-----------------------------------------------|--------------------------------------------------------------------------------------------|
| SIMPLE_SCHEDULER_API_URL                      | The URL to use for the API. e.g. `:8080`.                                                  |
| SIMPLE_SCHEDULER_DB_TYPE                      | The type of database to use. e.g. `mongodb`.                                               |
| SIMPLE_SCHEDULER_MESSAGEBUS_TYPE              | The type of message bus to use. e.g. `rabbitmq`.                                           |
| SIMPLE_SCHEDULER_DB_CONNECTION_STRING         | The connection string of the database.                                                     |
| SIMPLE_SCHEDULER_DB_NAME                      | The name of the database to connect to.                                                    |
| SIMPLE_SCHEDULER_MESSAGEBUS_CONNECTION_STRING | The connection string of the message bus.                                                  |
| SIMPLE_SCHEDULER_OIDC_ISSUER                  | The URL of the OIDC issuer to use. e.g. http://localhost:8080/realms/simple-scheduler      |

### CLI
This application allows you to manage jobs and runs in a terminal.

#### Running
The CLI makes calls to the [API](#api) so ensure that is running.

To run locally use
```bash
cd services/cli
go build .
./cli --help
```

#### Commands
See [simple-scheduler-cli](services/cli/docs/simple-scheduler-cli.md) for
documentation on commands.
